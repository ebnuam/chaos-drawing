// Drawing Chaos Online — Canvas Engine
// Dukungan mouse, pen, dan layar sentuh dengan koordinat proporsional (1200x800).
// Dilengkapi Undo, Redo, Eraser, Smoothing, dan ekspor goresan untuk WebSocket.

(function (window) {
  'use strict';

  const DEFAULT_COLORS = [
    { name: 'Tinta Hitam', value: '#1a1526' },
    { name: 'Merah Coral', value: '#ff4757' },
    { name: 'Merah Muda', value: '#ff6b81' },
    { name: 'Oranye Cerah', value: '#ffa502' },
    { name: 'Kuning Madu', value: '#ffc53d' },
    { name: 'Hijau Daun', value: '#2ed573' },
    { name: 'Biru Samudra', value: '#1e90ff' },
    { name: 'Biru Langit', value: '#70a1ff' },
    { name: 'Ungu Chaos', value: '#6c4cf1' },
    { name: 'Cokelat Kayu', value: '#8b572a' },
    { name: 'Putih Kertas', value: '#ffffff' }
  ];

  const DEFAULT_SIZES = [
    { name: 'Kecil', value: 4, iconSize: 6 },
    { name: 'Sedang', value: 8, iconSize: 10 },
    { name: 'Besar', value: 16, iconSize: 16 },
    { name: 'Ekstra', value: 28, iconSize: 22 }
  ];

  class DrawingBoard {
    constructor(canvasEl, options = {}) {
      this.canvas = canvasEl;
      this.ctx = this.canvas.getContext('2d', { willReadFrequently: false });

      this.width = options.width || 1200;
      this.height = options.height || 800;
      this.canvas.width = this.width;
      this.canvas.height = this.height;

      this.currentColor = DEFAULT_COLORS[0].value;
      this.currentSize = DEFAULT_SIZES[1].value;
      this.isEraser = false;
      this.readOnly = false;

      this.isDrawing = false;
      this.currentStroke = null;

      // History stacks untuk undo / redo
      this.strokes = [];
      this.redoStack = [];

      // Callbacks
      this.onHistoryChange = options.onHistoryChange || null;
      this.onStrokeComplete = options.onStrokeComplete || null;
      this.onClear = options.onClear || null;
      this.onUndo = options.onUndo || null;

      this.initBackground();
      this.bindEvents();
    }

    initBackground() {
      this.ctx.fillStyle = '#ffffff';
      this.ctx.fillRect(0, 0, this.width, this.height);
    }

    bindEvents() {
      // Pointer events menyatukan mouse, stylus/pen, dan touch
      this.canvas.addEventListener('pointerdown', (e) => this.handlePointerDown(e));
      this.canvas.addEventListener('pointermove', (e) => this.handlePointerMove(e));
      this.canvas.addEventListener('pointerup', (e) => this.handlePointerUp(e));
      this.canvas.addEventListener('pointercancel', (e) => this.handlePointerUp(e));
      this.canvas.addEventListener('pointerleave', (e) => {
        if (this.isDrawing) this.handlePointerUp(e);
      });

      // Mencegah scroll / pinch default saat menggambar di HP
      this.canvas.addEventListener('touchstart', (e) => e.preventDefault(), { passive: false });
      this.canvas.addEventListener('touchmove', (e) => e.preventDefault(), { passive: false });

      // Keyboard shortcuts: Ctrl+Z (Undo), Ctrl+Shift+Z / Ctrl+Y (Redo)
      window.addEventListener('keydown', (e) => {
        if (this.readOnly) return;
        const isInput = ['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName);
        if (isInput) return;

        if ((e.ctrlKey || e.metaKey) && !e.shiftKey && e.key.toLowerCase() === 'z') {
          e.preventDefault();
          this.undo();
        } else if ((e.ctrlKey || e.metaKey) && ((e.shiftKey && e.key.toLowerCase() === 'z') || e.key.toLowerCase() === 'y')) {
          e.preventDefault();
          this.redo();
        }
      });
    }

    getCanvasCoordinates(event) {
      const rect = this.canvas.getBoundingClientRect();
      const scaleX = this.width / rect.width;
      const scaleY = this.height / rect.height;

      const clientX = event.clientX;
      const clientY = event.clientY;

      const x = Math.round((clientX - rect.left) * scaleX);
      const y = Math.round((clientY - rect.top) * scaleY);

      return {
        x: Math.max(0, Math.min(this.width, x)),
        y: Math.max(0, Math.min(this.height, y))
      };
    }

    handlePointerDown(e) {
      if (this.readOnly) return;
      if (e.button !== undefined && e.button !== 0) return; // Hanya klik kiri

      this.canvas.setPointerCapture(e.pointerId);
      this.isDrawing = true;

      const pos = this.getCanvasCoordinates(e);

      this.currentStroke = {
        color: this.isEraser ? '#ffffff' : this.currentColor,
        size: this.currentSize,
        isEraser: this.isEraser,
        points: [pos]
      };

      // Buat titik tunggal langsung terlihat saat diklik
      this.drawPoint(pos, this.currentStroke.color, this.currentStroke.size);
    }

    handlePointerMove(e) {
      if (!this.isDrawing || !this.currentStroke || this.readOnly) return;

      const pos = this.getCanvasCoordinates(e);
      const pts = this.currentStroke.points;
      const lastPos = pts[pts.length - 1];

      // Abaikan jika pointer belum bergerak cukup jauh (optimasi performa)
      const dist = Math.hypot(pos.x - lastPos.x, pos.y - lastPos.y);
      if (dist < 2) return;

      pts.push(pos);
      this.drawSegment(lastPos, pos, this.currentStroke.color, this.currentStroke.size);
    }

    handlePointerUp(e) {
      if (!this.isDrawing) return;
      this.isDrawing = false;

      if (e.pointerId) {
        try {
          this.canvas.releasePointerCapture(e.pointerId);
        } catch (_) {}
      }

      if (this.currentStroke && this.currentStroke.points.length > 0) {
        this.strokes.push(this.currentStroke);
        this.redoStack = []; // Reset redo setelah goresan baru

        if (typeof this.onHistoryChange === 'function') {
          this.onHistoryChange({
            canUndo: this.canUndo(),
            canRedo: this.canRedo()
          });
        }

        if (typeof this.onStrokeComplete === 'function') {
          this.onStrokeComplete(this.currentStroke);
        }
      }

      this.currentStroke = null;
    }

    drawPoint(point, color, size) {
      this.ctx.fillStyle = color;
      this.ctx.beginPath();
      this.ctx.arc(point.x, point.y, size / 2, 0, Math.PI * 2);
      this.ctx.fill();
    }

    drawSegment(p1, p2, color, size) {
      this.ctx.save();
      this.ctx.strokeStyle = color;
      this.ctx.lineWidth = size;
      this.ctx.lineCap = 'round';
      this.ctx.lineJoin = 'round';

      this.ctx.beginPath();
      this.ctx.moveTo(p1.x, p1.y);
      this.ctx.lineTo(p2.x, p2.y);
      this.ctx.stroke();
      this.ctx.restore();
    }

    drawFullStroke(stroke) {
      if (!stroke || !stroke.points || stroke.points.length === 0) return;

      const pts = stroke.points;
      if (pts.length === 1) {
        this.drawPoint(pts[0], stroke.color, stroke.size);
        return;
      }

      this.ctx.save();
      this.ctx.strokeStyle = stroke.color;
      this.ctx.lineWidth = stroke.size;
      this.ctx.lineCap = 'round';
      this.ctx.lineJoin = 'round';

      this.ctx.beginPath();
      this.ctx.moveTo(pts[0].x, pts[0].y);

      for (let i = 1; i < pts.length; i++) {
        const midX = (pts[i - 1].x + pts[i].x) / 2;
        const midY = (pts[i - 1].y + pts[i].y) / 2;
        this.ctx.quadraticCurveTo(pts[i - 1].x, pts[i - 1].y, midX, midY);
      }

      this.ctx.lineTo(pts[pts.length - 1].x, pts[pts.length - 1].y);
      this.ctx.stroke();
      this.ctx.restore();
    }

    redrawAll() {
      this.initBackground();
      for (const stroke of this.strokes) {
        this.drawFullStroke(stroke);
      }
    }

    // ================= ACTIONS =================

    canUndo() {
      return this.strokes.length > 0;
    }

    canRedo() {
      return this.redoStack.length > 0;
    }

    undo() {
      if (!this.canUndo() || this.readOnly) return false;
      const stroke = this.strokes.pop();
      this.redoStack.push(stroke);
      this.redrawAll();

      if (typeof this.onHistoryChange === 'function') {
        this.onHistoryChange({ canUndo: this.canUndo(), canRedo: this.canRedo() });
      }
      if (typeof this.onUndo === 'function') {
        this.onUndo();
      }
      return true;
    }

    redo() {
      if (!this.canRedo() || this.readOnly) return false;
      const stroke = this.redoStack.pop();
      this.strokes.push(stroke);
      this.drawFullStroke(stroke);

      if (typeof this.onHistoryChange === 'function') {
        this.onHistoryChange({ canUndo: this.canUndo(), canRedo: this.canRedo() });
      }
      if (typeof this.onStrokeComplete === 'function') {
        this.onStrokeComplete(stroke);
      }
      return true;
    }

    clear(notify = true) {
      if (this.readOnly) return;
      this.strokes = [];
      this.redoStack = [];
      this.initBackground();

      if (notify && typeof this.onHistoryChange === 'function') {
        this.onHistoryChange({ canUndo: false, canRedo: false });
      }
      if (notify && typeof this.onClear === 'function') {
        this.onClear();
      }
    }

    applyRemoteStroke(stroke) {
      if (!stroke) return;
      this.strokes.push(stroke);
      this.drawFullStroke(stroke);
    }

    applyRemoteClear() {
      this.strokes = [];
      this.redoStack = [];
      this.initBackground();
      if (typeof this.onHistoryChange === 'function') {
        this.onHistoryChange({ canUndo: false, canRedo: false });
      }
    }

    applyRemoteUndo() {
      if (this.strokes.length > 0) {
        this.strokes.pop();
        this.redrawAll();
        if (typeof this.onHistoryChange === 'function') {
          this.onHistoryChange({ canUndo: this.canUndo(), canRedo: this.canRedo() });
        }
      }
    }

    setColor(color) {
      this.currentColor = color;
      this.isEraser = false;
    }

    setSize(size) {
      this.currentSize = size;
    }

    setEraser(active) {
      this.isEraser = !!active;
    }

    setReadOnly(readOnly) {
      this.readOnly = !!readOnly;
      this.canvas.style.cursor = this.readOnly ? 'default' : 'crosshair';
    }

    saveAsPNG(filename = 'drawing-chaos.png') {
      const link = document.createElement('a');
      link.download = filename;
      link.href = this.canvas.toDataURL('image/png');
      link.click();
    }

    // Ekspor seluruh goresan untuk pemulihan state saat pemain baru bergabung
    exportStrokes() {
      return JSON.parse(JSON.stringify(this.strokes));
    }

    // Impor goresan dari room server
    importStrokes(strokes) {
      this.strokes = Array.isArray(strokes) ? strokes : [];
      this.redoStack = [];
      this.redrawAll();
      if (typeof this.onHistoryChange === 'function') {
        this.onHistoryChange({ canUndo: this.canUndo(), canRedo: this.canRedo() });
      }
    }
  }

  // Helper untuk merender toolbar alat gambar ke DOM
  DrawingBoard.setupToolbar = function (board, elements = {}) {
    const { colorsEl, sizesEl, eraserBtn, undoBtn, redoBtn, clearBtn, saveBtn, confirmDialog } = elements;

    // 1. Render Warna
    if (colorsEl) {
      colorsEl.innerHTML = '';
      DEFAULT_COLORS.forEach((c, idx) => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = `color-swatch ${idx === 0 ? 'active' : ''}`;
        btn.style.setProperty('--swatch-color', c.value);
        btn.title = c.name;
        btn.setAttribute('aria-label', c.name);

        btn.addEventListener('click', () => {
          board.setColor(c.value);
          colorsEl.querySelectorAll('.color-swatch').forEach(el => el.classList.remove('active'));
          btn.classList.add('active');
          if (eraserBtn) eraserBtn.setAttribute('aria-pressed', 'false');
          if (window.SoundManager) window.SoundManager.playClick();
        });

        colorsEl.appendChild(btn);
      });
    }

    // 2. Render Ukuran Goresan
    if (sizesEl) {
      sizesEl.innerHTML = '';
      DEFAULT_SIZES.forEach((s, idx) => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = `size-btn ${idx === 1 ? 'active' : ''}`;
        btn.title = `Ukuran ${s.name} (${s.value}px)`;
        btn.setAttribute('aria-label', s.name);

        const dot = document.createElement('span');
        dot.className = 'size-dot';
        dot.style.width = `${s.iconSize}px`;
        dot.style.height = `${s.iconSize}px`;
        btn.appendChild(dot);

        btn.addEventListener('click', () => {
          board.setSize(s.value);
          sizesEl.querySelectorAll('.size-btn').forEach(el => el.classList.remove('active'));
          btn.classList.add('active');
          if (window.SoundManager) window.SoundManager.playClick();
        });

        sizesEl.appendChild(btn);
      });
    }

    // 3. Tombol Penghapus
    if (eraserBtn) {
      eraserBtn.addEventListener('click', () => {
        const isActive = eraserBtn.getAttribute('aria-pressed') === 'true';
        const next = !isActive;
        eraserBtn.setAttribute('aria-pressed', next ? 'true' : 'false');
        board.setEraser(next);

        if (next && colorsEl) {
          colorsEl.querySelectorAll('.color-swatch').forEach(el => el.classList.remove('active'));
        }
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    // 4. Undo & Redo
    board.onHistoryChange = ({ canUndo, canRedo }) => {
      if (undoBtn) undoBtn.disabled = !canUndo;
      if (redoBtn) redoBtn.disabled = !canRedo;
    };

    if (undoBtn) {
      undoBtn.addEventListener('click', () => {
        board.undo();
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    if (redoBtn) {
      redoBtn.addEventListener('click', () => {
        board.redo();
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    // 5. Bersihkan Kanvas
    if (clearBtn) {
      clearBtn.addEventListener('click', () => {
        if (confirmDialog && typeof confirmDialog.showModal === 'function') {
          confirmDialog.showModal();
        } else if (window.confirm('Hapus seluruh kanvas?')) {
          board.clear();
        }
      });
    }

    if (confirmDialog) {
      confirmDialog.addEventListener('close', () => {
        if (confirmDialog.returnValue === 'confirm') {
          board.clear();
          if (window.SoundManager) window.SoundManager.playPop();
        }
      });
    }

    // 6. Unduh Gambar
    if (saveBtn) {
      saveBtn.addEventListener('click', () => {
        board.saveAsPNG();
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }
  };

  window.DrawingBoard = DrawingBoard;
})(window);

