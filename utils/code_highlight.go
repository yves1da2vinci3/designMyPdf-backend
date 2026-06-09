package utils

// Code highlight via highlight.js CDN (mirrors frontend/utils/codeHighlightShell.ts).

const codeHighlightCDNVersion = "11.11.1"
const codeHighlightTheme = "github-dark"

// CodeHighlightFitBodyClass is set on preview + export HTML body for code wrap layout.
const CodeHighlightFitBodyClass = "dmp-code-fit"

// CodeHighlightPdfBodyClass is an alias for CodeHighlightFitBodyClass.
const CodeHighlightPdfBodyClass = CodeHighlightFitBodyClass

func codeHighlightCDNBase() string {
	return "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/" + codeHighlightCDNVersion
}

// CodeHighlightHeadTags returns link + script tags for hljs bundle and theme.
func CodeHighlightHeadTags() string {
	base := codeHighlightCDNBase()
	return `<link rel="stylesheet" href="` + base + `/styles/` + codeHighlightTheme + `.min.css" crossorigin="anonymous" referrerpolicy="no-referrer" />
<script src="` + base + `/highlight.min.js" crossorigin="anonymous" referrerpolicy="no-referrer"></script>`
}

// CodeHighlightPreviewCSS styles pre/code for preview iframes (horizontal scroll).
func CodeHighlightPreviewCSS() string {
	return codeHighlightContrastCSS() + `
*:has(>pre>code[class*="language-"]),*:has(>pre>code.hljs){overflow-x:auto!important;overflow-y:visible!important;max-width:100%!important}
pre.hljs-wrap,pre.hljs,pre:has(>code[class*="language-"]),pre:has(>code.hljs){margin:.75rem 0;padding:0;max-width:100%!important;width:100%!important;box-sizing:border-box!important;overflow-x:auto!important;overflow-y:visible!important;border-radius:6px;background:#0d1117;-webkit-print-color-adjust:exact;print-color-adjust:exact}
pre.hljs-wrap code,pre>code.hljs,pre>code[class*="language-"]{display:block;box-sizing:border-box;padding:.875rem 1rem;font-size:.75rem;line-height:1.55;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;white-space:pre;width:max-content;min-width:100%;max-width:none;-webkit-print-color-adjust:exact;print-color-adjust:exact}`
}

// CodeHighlightPdfFitCSS wraps code inside blocks for PDF export (no scroll).
func CodeHighlightPdfFitCSS() string {
	body := CodeHighlightFitBodyClass
	return codeHighlightContrastCSS() + `
body.` + body + ` *:has(>pre>code[class*="language-"]),body.` + body + ` *:has(>pre>code.hljs){overflow:visible!important;overflow-x:visible!important;max-width:100%!important;width:100%!important;box-sizing:border-box!important}
body.` + body + ` pre.hljs-wrap,body.` + body + ` pre.hljs,body.` + body + ` pre:has(>code[class*="language-"]),body.` + body + ` pre:has(>code.hljs){margin:.75rem 0;padding:0;max-width:100%!important;width:100%!important;box-sizing:border-box!important;overflow:visible!important;overflow-x:visible!important;border-radius:6px;background:#0d1117;-webkit-print-color-adjust:exact;print-color-adjust:exact}
body.` + body + ` pre>code.hljs,body.` + body + ` pre>code[class*="language-"]{display:block;box-sizing:border-box;padding:.75rem .875rem;font-size:.65rem;line-height:1.5;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;white-space:pre-wrap!important;word-break:break-word!important;overflow-wrap:anywhere!important;width:100%!important;max-width:100%!important;min-width:0!important;-webkit-print-color-adjust:exact;print-color-adjust:exact}
body.` + body + ` pre>code.hljs span,body.` + body + ` pre>code.hljs *{white-space:inherit!important;word-break:inherit!important;overflow-wrap:inherit!important}
@media print{body.` + body + ` pre>code.hljs,body.` + body + ` pre>code[class*="language-"]{white-space:pre-wrap!important;word-break:break-word!important;overflow-wrap:anywhere!important}}`
}

// CodeHighlightBaseCSS is an alias for preview styles.
func CodeHighlightBaseCSS() string {
	return CodeHighlightPreviewCSS()
}

func codeHighlightContrastCSS() string {
	return `pre>code.hljs,pre>code[class*="language-"]{background:#0d1117!important;color:#e6edf3!important}
pre>code.hljs .hljs-comment,pre>code.hljs .hljs-quote{color:#8b949e!important}
pre>code.hljs .hljs-keyword,pre>code.hljs .hljs-selector-tag,pre>code.hljs .hljs-literal{color:#ff7b72!important}
pre>code.hljs .hljs-string,pre>code.hljs .hljs-doctag,pre>code.hljs .hljs-regexp{color:#a5d6ff!important}
pre>code.hljs .hljs-title,pre>code.hljs .hljs-section,pre>code.hljs .hljs-type,pre>code.hljs .hljs-built_in{color:#d2a8ff!important}
pre>code.hljs .hljs-function,pre>code.hljs .hljs-title.function_{color:#d2a8ff!important}
pre>code.hljs .hljs-attr,pre>code.hljs .hljs-attribute,pre>code.hljs .hljs-variable{color:#79c0ff!important}
pre>code.hljs .hljs-number,pre>code.hljs .hljs-symbol{color:#79c0ff!important}
pre>code.hljs .hljs-meta{color:#8b949e!important}
pre>code.hljs .hljs-params{color:#e6edf3!important}
pre>code.hljs .hljs-name{color:#7ee787!important}`
}

const codeHighlightFitBlocksJS = `
function fitCodeBlocksForPdf() {
  var sel = 'pre > code.hljs, pre > code[class*="language-"]';
  document.querySelectorAll(sel).forEach(function(code) {
    code.style.setProperty('white-space', 'pre-wrap', 'important');
    code.style.setProperty('word-break', 'break-word', 'important');
    code.style.setProperty('overflow-wrap', 'anywhere', 'important');
    code.style.setProperty('width', '100%', 'important');
    code.style.setProperty('max-width', '100%', 'important');
    code.style.setProperty('min-width', '0', 'important');
    code.querySelectorAll('span').forEach(function(span) {
      span.style.setProperty('white-space', 'inherit', 'important');
      span.style.setProperty('word-break', 'inherit', 'important');
      span.style.setProperty('overflow-wrap', 'inherit', 'important');
    });
    var pre = code.parentElement;
    if (pre && pre.tagName === 'PRE') {
      pre.style.setProperty('overflow', 'visible', 'important');
      pre.style.setProperty('overflow-x', 'visible', 'important');
      pre.style.setProperty('max-width', '100%', 'important');
      pre.style.setProperty('width', '100%', 'important');
    }
  });
  document.querySelectorAll('*:has(> pre > code)').forEach(function(el) {
    el.style.setProperty('overflow-x', 'visible', 'important');
    el.style.setProperty('max-width', '100%', 'important');
  });
}`

// CodeHighlightPdfAwaitJS highlights code and forces wrap for PDF export.
func CodeHighlightPdfAwaitJS() string {
	return `(function() {
  ` + codeHighlightFitBlocksJS + `
  return new Promise(function(resolve) {
    function done() {
      fitCodeBlocksForPdf();
      setTimeout(resolve, 100);
    }
    if (typeof window.hljs === 'undefined' || !window.hljs.highlightAll) {
      fitCodeBlocksForPdf();
      done();
      return;
    }
    try { window.hljs.highlightAll(); } catch (e) { console.warn('[code-highlight]', e); }
    fitCodeBlocksForPdf();
    done();
  });
})()`
}

// CodeHighlightAwaitJS runs highlightAll and resolves after a short paint delay (chromedp.Evaluate).
func CodeHighlightAwaitJS() string {
	return CodeHighlightPdfAwaitJS()
}
