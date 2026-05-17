package utils

import (
	_ "embed"
	"fmt"
)

//go:embed pdf_pagination.js
var pdfPaginationJS string

// ApplyPdfPageBreakHintsJS returns a chromedp.Evaluate expression that runs pagination hints in-page.
// The pagination script must be loaded in the document first (PaginationScriptTag).
func ApplyPdfPageBreakHintsJS(contentAreaHeight int, threshold float64, resetExisting bool) string {
	return fmt.Sprintf(
		`window.__applyPdfPageBreakHints({cah:%d,threshold:%g,resetExisting:%t});`,
		contentAreaHeight,
		threshold,
		resetExisting,
	)
}

// PaginationScriptTag returns a script tag to define __applyPdfPageBreakHints before use.
func PaginationScriptTag() string {
	return "<script>" + pdfPaginationJS + "</script>"
}

// OrphanThresholdPx returns the orphan/widow threshold in pixels for a content area height.
func OrphanThresholdPx(contentAreaHeight int) float64 {
	return float64(contentAreaHeight) * OrphanThreshold
}

// WaitForTailwindJS returns a Promise that resolves when Tailwind CDN styles apply (max ~2.5s).
func WaitForTailwindJS() string {
	return `(function() {
  return new Promise(function(resolve) {
    var start = Date.now();
    function check() {
      var testDiv = document.createElement('div');
      testDiv.className = 'bg-blue-500';
      document.body.appendChild(testDiv);
      var ok = window.getComputedStyle(testDiv).backgroundColor !== 'rgba(0, 0, 0, 0)';
      document.body.removeChild(testDiv);
      if (ok || Date.now() - start > 2500) resolve();
      else setTimeout(check, 80);
    }
    check();
  });
})()`
}
