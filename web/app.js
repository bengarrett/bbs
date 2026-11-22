(() => {
  const go = new Go();
  let wasmReady = false;
  let currentFileBytes = null;
  let currentFileName = 'bbs';
  let activeBlobUrl = null;
  let lastTextOutput = '';

  const bbsForm = document.getElementById('bbsForm');
  const wasmStatus = document.getElementById('wasmStatus');
  const filePicker = document.getElementById('filePicker');
  const dropzone = document.getElementById('dropzone');
  const selectedFileInfo = document.getElementById('selectedFileInfo');
  const statusMessage = document.getElementById('statusMessage');

  const formatSelect = document.getElementById('format');
  const bbsKindSelect = document.getElementById('bbsKind');
  const utf8Checkbox = document.getElementById('utf8');
  const cssCheckbox = document.getElementById('css');
  const unlockCheckbox = document.getElementById('unlock');

  const previewContainer = document.getElementById('previewContainer');
  const previewPlaceholder = previewContainer.querySelector(
    '.preview-placeholder'
  );
  const previewOutput = document.getElementById('previewOutput');
  const previewImage = document.getElementById('previewImage');
  const resultMeta = document.getElementById('resultMeta');
  const metaDetected = document.getElementById('metaDetected');
  const metaInputSize = document.getElementById('metaInputSize');
  const metaOutputSize = document.getElementById('metaOutputSize');
  const metaTime = document.getElementById('metaTime');
  const downloadBtn = document.getElementById('downloadBtn');
  const copyBtn = document.getElementById('copyBtn');
  const sampleLink = document.getElementById('sampleLink');

  function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  }

  function syncControls() {
    const isHTML = formatSelect.value === 'html';
    const isPNG = formatSelect.value === 'png';
    cssCheckbox.disabled = !isHTML;
    utf8Checkbox.disabled = isHTML || isPNG;
  }

  function applyDynamicCSS(kind = 'pcboard') {
    if (!wasmReady || typeof globalThis.bbs?.css !== 'function') return;
    let styleEl = document.getElementById('bbsDynamicStyle');
    if (!styleEl && document.head) {
      styleEl = document.createElement('style');
      styleEl.id = 'bbsDynamicStyle';
      document.head.appendChild(styleEl);
    }
    if (styleEl) {
      styleEl.textContent = globalThis.bbs.css(kind);
    }
  }

  async function initWasm() {
    try {
      const response = await fetch('bbs-web.wasm');
      if (!response.ok) {
        throw new Error(
          `Failed to fetch bbs-web.wasm (${response.status} ${response.statusText})`
        );
      }
      const { instance } = await WebAssembly.instantiateStreaming(
        response,
        go.importObject
      );
      go.run(instance);
      wasmReady = true;
      applyDynamicCSS('pcboard');
      wasmStatus.textContent = 'Ready v' + globalThis.bbs.version();
      if (currentFileBytes) triggerConvert();
    } catch (err) {
      console.error('WASM load error:', err);
      wasmStatus.textContent = 'WASM Failed';
      showStatus('Could not load WebAssembly module: ' + err.message, 'error');
    }
  }

  function showStatus(msg, type = 'error') {
    statusMessage.textContent = msg;
    statusMessage.className = `status-message ${type}`;
    statusMessage.classList.remove('hidden');
  }

  function clearStatus() {
    statusMessage.classList.add('hidden');
    statusMessage.textContent = '';
  }

  function handleFile(file) {
    clearStatus();
    currentFileName = file.name.replace(/\.[^/.]+$/, '') || 'bbs';
    const reader = new FileReader();
    reader.onload = (e) => {
      currentFileBytes = new Uint8Array(e.target.result);
      selectedFileInfo.textContent = `Selected: ${file.name} (${formatBytes(currentFileBytes.length)})`;
      selectedFileInfo.classList.remove('hidden');
      if (wasmReady) {
        triggerConvert();
      }
    };
    reader.onerror = () => showStatus('Failed to read file', 'error');
    reader.readAsArrayBuffer(file);
  }

  dropzone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropzone.classList.add('dragover');
  });
  dropzone.addEventListener('dragleave', () =>
    dropzone.classList.remove('dragover')
  );
  dropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropzone.classList.remove('dragover');
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFile(e.dataTransfer.files[0]);
    }
  });
  filePicker.addEventListener('change', (e) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFile(e.target.files[0]);
    }
  });

  if (sampleLink) {
    sampleLink.addEventListener('click', (e) => {
      e.preventDefault();
      clearStatus();
      // Built-in CP437 PCBoard sample (0xC9/0xCD/0xBB box + @X colors)
      const lines = [
        '@CLS@@X1F',
        '\xC9' + '\xCD'.repeat(35) + '\xBB',
        '\xBA  @X1EWELCOME TO THE BBS TEST SYSTEM!  @X1F\xBA',
        '\xBA  @X1BPCBoard @X color codes -> UTF-8  @X1F\xBA',
        '\xC8' + '\xCD'.repeat(35) + '\xBC@X07',
        '@X0A  Online 24/7 @X07| @X0CSysop: Ben @X07| @X0E28.8k',
      ];
      const text = lines.join('\r\n');
      const bytes = new Uint8Array(text.length);
      for (let i = 0; i < text.length; i++) {
        bytes[i] = text.charCodeAt(i) & 0xff;
      }
      currentFileName = 'sample';
      currentFileBytes = bytes;
      selectedFileInfo.textContent = `Selected: sample.pcb (${formatBytes(bytes.length)})`;
      selectedFileInfo.classList.remove('hidden');
      if (wasmReady) triggerConvert();
    });
  }

  async function triggerConvert() {
    if (!wasmReady || !currentFileBytes) return;

    clearStatus();
    const format = formatSelect.value;
    const opts = {
      bbs: bbsKindSelect.value,
      utf8: utf8Checkbox.checked,
      css: cssCheckbox.checked,
      unlock: unlockCheckbox.checked,
    };

    const startTime = performance.now();

    try {
      const info = globalThis.bbs.detect(currentFileBytes);
      metaDetected.textContent = info.valid ? info.format : 'None detected';

      let outBytes;
      let ext = 'txt';
      let mime = 'text/plain;charset=utf-8';

      if (format === 'html') {
        const activeKind =
          opts.bbs === 'auto' ? info.name || 'pcboard' : opts.bbs;
        applyDynamicCSS(activeKind);
        const html = await globalThis.bbs.toHTML(currentFileBytes, opts);
        lastTextOutput = html;
        previewImage.classList.add('hidden');
        previewImage.removeAttribute('src');
        previewOutput.innerHTML = await globalThis.bbs.toHTML(
          currentFileBytes,
          {
            bbs: opts.bbs,
            css: false,
            unlock: opts.unlock,
          }
        );
        previewOutput.classList.remove('hidden');
        copyBtn.classList.remove('hidden');
        outBytes = new TextEncoder().encode(html);
        ext = 'html';
        mime = 'text/html;charset=utf-8';
      } else if (format === 'png') {
        outBytes = await globalThis.bbs.toPNG(currentFileBytes, opts);
        lastTextOutput = '';
        previewOutput.classList.add('hidden');
        previewOutput.textContent = '';
        copyBtn.classList.add('hidden');
        ext = 'png';
        mime = 'image/png';
      } else {
        const isANSI = format === 'ansi';
        const convert = isANSI ? globalThis.bbs.toANSI : globalThis.bbs.strip;
        const rawOut = await convert(currentFileBytes, opts);
        const previewText = opts.utf8
          ? rawOut
          : await convert(currentFileBytes, {
              bbs: opts.bbs,
              utf8: true,
              unlock: opts.unlock,
            });
        lastTextOutput = previewText;
        previewImage.classList.add('hidden');
        previewImage.removeAttribute('src');
        previewOutput.textContent = previewText;
        previewOutput.classList.remove('hidden');
        copyBtn.classList.remove('hidden');
        outBytes =
          typeof rawOut === 'string'
            ? new TextEncoder().encode(rawOut)
            : rawOut;
        ext = isANSI ? 'ans' : 'txt';
        if (isANSI) {
          mime = 'application/octet-stream';
        }
      }

      const elapsed = ((performance.now() - startTime) / 1000).toFixed(3);

      if (activeBlobUrl) {
        URL.revokeObjectURL(activeBlobUrl);
      }
      activeBlobUrl = URL.createObjectURL(new Blob([outBytes], { type: mime }));

      if (format === 'png') {
        previewImage.src = activeBlobUrl;
        previewImage.classList.remove('hidden');
      }

      previewContainer.classList.remove('empty');
      previewPlaceholder.classList.add('hidden');

      metaInputSize.textContent = formatBytes(currentFileBytes.length);
      metaOutputSize.textContent = formatBytes(outBytes.length);
      metaTime.textContent = `${elapsed}s`;

      downloadBtn.href = activeBlobUrl;
      downloadBtn.download = `${currentFileName}.${ext}`;
      downloadBtn.textContent = `Download .${ext}`;
      resultMeta.classList.remove('hidden');
    } catch (err) {
      showStatus('Conversion error: ' + err.message, 'error');
    }
  }

  copyBtn.addEventListener('click', async () => {
    if (!lastTextOutput) return;
    try {
      await navigator.clipboard.writeText(lastTextOutput);
      const prev = copyBtn.textContent;
      copyBtn.textContent = 'Copied!';
      setTimeout(() => {
        copyBtn.textContent = prev;
      }, 1500);
    } catch (err) {
      showStatus('Copy failed: ' + err.message, 'error');
    }
  });

  bbsForm.addEventListener('submit', (e) => e.preventDefault());

  bbsForm.addEventListener('change', () => {
    syncControls();
    if (currentFileBytes && wasmReady) {
      triggerConvert();
    }
  });

  bbsForm.addEventListener('reset', () => {
    clearStatus();
    currentFileBytes = null;
    currentFileName = 'bbs';
    lastTextOutput = '';

    selectedFileInfo.classList.add('hidden');
    selectedFileInfo.textContent = '';

    if (activeBlobUrl) {
      URL.revokeObjectURL(activeBlobUrl);
      activeBlobUrl = null;
    }

    previewContainer.classList.add('empty');
    previewOutput.innerHTML = '';
    previewOutput.classList.add('hidden');
    previewImage.removeAttribute('src');
    previewImage.classList.add('hidden');
    copyBtn.classList.remove('hidden');
    previewPlaceholder.classList.remove('hidden');
    resultMeta.classList.add('hidden');
    downloadBtn.removeAttribute('href');
    setTimeout(syncControls, 0);
  });

  syncControls();
  initWasm();
})();
