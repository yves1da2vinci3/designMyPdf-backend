// Keep in sync with frontend/utils/pdfPagination.ts (pageBreakHintsEvaluate).
window.__applyPdfPageBreakHints = function (params) {
  var cah = params.cah;
  var threshold = params.threshold;
  var resetExisting = params.resetExisting;

  function collectBlocks(container) {
    var direct = Array.from(container.children);
    if (direct.length === 1) {
      var nested = Array.from(direct[0].children);
      if (nested.length > 0) return nested;
    }
    return direct.length > 0
      ? direct
      : Array.from(
          container.querySelectorAll(
            'section,article,table,[data-pdf-block],.pdf-keep-together',
          ),
        );
  }

  function shouldAvoid(el) {
    if (el.tagName !== 'TABLE' && !el.querySelector('canvas[data-chart-type]')) {
      return false;
    }
    var h = el.getBoundingClientRect().height;
    return h > 0 && h <= cah;
  }

  var container = document.querySelector('.content');
  if (!container) return;

  if (resetExisting) {
    container.querySelectorAll('.pdf-page-break-before').forEach(function (el) {
      el.style.removeProperty('break-before');
      el.style.removeProperty('page-break-before');
      el.classList.remove('pdf-page-break-before');
    });
    container.querySelectorAll('.pdf-avoid-break-inside').forEach(function (el) {
      el.classList.remove('pdf-avoid-break-inside');
    });
  }

  var blocks = collectBlocks(container);
  var containerTop = container.getBoundingClientRect().top;
  var lastBreakPage = -1;

  for (var i = 0; i < blocks.length; i++) {
    var el = blocks[i];
    el.classList.remove('pdf-avoid-break-inside');
    if (shouldAvoid(el)) {
      el.classList.add('pdf-avoid-break-inside');
    }
  }

  for (var j = 0; j < blocks.length; j++) {
    var block = blocks[j];
    var rect = block.getBoundingClientRect();
    if (rect.height <= 0) continue;

    var blockTop = rect.top - containerTop;
    var pageNum = Math.floor(blockTop / cah);
    var posOnPage = blockTop % cah;
    var remaining = cah - posOnPage;

    var needsBreak = false;

    if (remaining > 0 && remaining < cah && remaining < threshold && pageNum !== lastBreakPage) {
      needsBreak = true;
    }

    if (!needsBreak && rect.height <= cah) {
      var blockEnd = blockTop + rect.height;
      var pageEnd = (pageNum + 1) * cah;
      if (blockEnd > pageEnd) {
        var overflowTail = blockEnd - pageEnd;
        if (overflowTail > 0 && overflowTail < threshold && pageNum !== lastBreakPage) {
          needsBreak = true;
        }
      }
    }

    if (needsBreak) {
      block.classList.add('pdf-page-break-before');
      block.style.breakBefore = 'page';
      block.style.pageBreakBefore = 'always';
      lastBreakPage = pageNum;
    }
  }
};
