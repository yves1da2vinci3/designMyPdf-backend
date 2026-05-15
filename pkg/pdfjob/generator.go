package pdfjob

import (
	"context"
	"crypto/md5"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/key"
	"designmypdf/pkg/storage"
	"designmypdf/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"net/url"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

type pdfCache struct {
	cache    map[string]string
	mu       sync.RWMutex
	maxItems int
}

var (
	pdfCacheInstance = &pdfCache{
		cache:    make(map[string]string),
		maxItems: 100,
	}
	storageInstance *storage.BackblazeStorage
	storageMu       sync.Mutex
)

func generateHash(templateContent string, data map[string]interface{}, format string, bgColor string, contentPadding string) string {
	dataBytes, _ := json.Marshal(data)
	hash := md5.Sum([]byte(templateContent + string(dataBytes) + format + bgColor + contentPadding))
	return hex.EncodeToString(hash[:])
}

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func isValidHex(color string) bool {
	return hexColorRe.MatchString(color)
}

func getStorageInstance() (*storage.BackblazeStorage, error) {
	storageMu.Lock()
	defer storageMu.Unlock()
	if storageInstance != nil {
		return storageInstance, nil
	}
	keyID, appKey, bucketName, ok := storage.B2ConfigFromEnv()
	if !ok {
		return nil, fmt.Errorf("backblaze B2 env missing: set BACKBLAZE_KEY_ID, BACKBLAZE_APP_KEY, BACKBLAZE_BUCKET_NAME")
	}
	b2, err := storage.NewBackblazeStorage(keyID, appKey, bucketName)
	if err != nil {
		return nil, err
	}
	storageInstance = b2
	return storageInstance, nil
}

// GeneratePdfForKey renders a PDF for the given key/template/data and returns
// the Backblaze URL. It is safe for concurrent use and works for both the
// synchronous HTTP handler and the async worker.
func GeneratePdfForKey(
	ctx context.Context,
	keyEntity *entities.Key,
	templateEntity *entities.Template,
	data map[string]interface{},
	format string,
) (string, error) {
	contentHash := generateHash(templateEntity.Content, data, format, templateEntity.PdfBackgroundColor, templateEntity.PdfContentPadding)

	pdfCacheInstance.mu.RLock()
	cachedURL, found := pdfCacheInstance.cache[contentHash]
	pdfCacheInstance.mu.RUnlock()
	if found {
		fmt.Printf("PDF found in cache: %s\n", cachedURL)
		go func() {
			svc := key.NewService(key.Repository{})
			if err := svc.IncreaseUsageCount(keyEntity.ID); err != nil {
				fmt.Printf("warning: failed to increase usage count: %v\n", err)
			}
		}()
		return cachedURL, nil
	}

	renderedHTML, err := utils.RenderTemplate(templateEntity.Content, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// Tailwind v4 Play CDN: browser script detects arbitrary classes at runtime.
	// Bootstrap uses a plain CSS CDN link (no JS DOM scan needed).
	var frameworkTag string
	switch templateEntity.Framework {
	case entities.Bootstrap:
		frameworkTag = `<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5/dist/css/bootstrap.min.css">`
	default:
		frameworkTag = `<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>`
	}

	fontImports := utils.ImportFontCreation(templateEntity.Fonts)
	fontCSS := utils.FontCssCreation(templateEntity.Fonts)

	var bgStyle string
	if isValidHex(templateEntity.PdfBackgroundColor) {
		bgStyle = fmt.Sprintf("body{background-color:%s!important}", templateEntity.PdfBackgroundColor)
	}

	pad := utils.EffectivePdfContentPadding(templateEntity.PdfContentPadding)
	padStyle := fmt.Sprintf(".content{box-sizing:border-box;padding:%s}", pad)

	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Preview</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    %s
    %s
    <style>%s %s %s</style>
</head>
<body class="overflow-x-hidden overflow-y-auto">
    <div class="content">%s</div>
</body>
</html>`, frameworkTag, fontImports, fontCSS, bgStyle, padStyle, renderedHTML)

	if err := os.MkdirAll("./uploads/template", 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	id := uuid.New()
	outputPath := fmt.Sprintf("./uploads/template/template_%s.pdf", id.String())

	runtime.GC()

	f, err := utils.GetFormat(format)
	if err != nil {
		return "", fmt.Errorf("invalid format %q: %w", format, err)
	}

	pool := GetBrowserPool()
	tabCtx, cancelTab := pool.NewTab()
	defer cancelTab()

	// Apply a 20s deadline on the tab so a hung render never blocks forever.
	tabCtx, cancelTimeout := context.WithTimeout(tabCtx, 20*time.Second)
	defer cancelTimeout()

	var pdfBuf []byte
	if err := chromedp.Run(tabCtx,
		chromedp.Navigate("data:text/html,"+url.PathEscape(fullHTML)),
		// Give Tailwind v4 browser script time to scan classes and inject CSS.
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`new Promise(r => setTimeout(r, 150))`, nil).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var runErr error
			pdfBuf, _, runErr = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(f.Width).
				WithPaperHeight(f.Height).
				WithMarginTop(0.0).
				WithMarginBottom(0.0).
				WithMarginLeft(0.0).
				WithMarginRight(0.0).
				Do(ctx)
			return runErr
		}),
	); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	if err := utils.SavePDF(outputPath, pdfBuf); err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	b2Storage, err := getStorageInstance()
	if err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("failed to initialize storage: %w", err)
	}

	storagePath := fmt.Sprintf("templates/%s.pdf", id.String())

	var wg sync.WaitGroup
	var uploadedURL string
	var uploadErr, countErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		uploadedURL, uploadErr = b2Storage.UploadFile(ctx, outputPath, storagePath)
	}()
	go func() {
		defer wg.Done()
		svc := key.NewService(key.Repository{})
		countErr = svc.IncreaseUsageCount(keyEntity.ID)
	}()
	wg.Wait()

	if uploadErr != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("failed to upload PDF: %w", uploadErr)
	}
	if countErr != nil {
		fmt.Printf("warning: failed to increase usage count: %v\n", countErr)
	}

	pdfCacheInstance.mu.Lock()
	if len(pdfCacheInstance.cache) >= pdfCacheInstance.maxItems {
		for k := range pdfCacheInstance.cache {
			delete(pdfCacheInstance.cache, k)
			break
		}
	}
	pdfCacheInstance.cache[contentHash] = uploadedURL
	pdfCacheInstance.mu.Unlock()

	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := os.Remove(outputPath); err != nil {
			fmt.Printf("warning: failed to delete local PDF: %v\n", err)
		}
	}()

	return uploadedURL, nil
}
