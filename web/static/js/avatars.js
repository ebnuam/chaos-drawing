// Drawing Chaos Online — Avatar Registry
// Kumpulan avatar karakter orisinal berbasis SVG inline tanpa ketergantungan aset luar.

(function (window) {
  'use strict';

  const AVATARS = [
    {
      id: 'cat-orange',
      name: 'Kucing Oyen',
      bg: '#ffe3b3',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <circle cx="32" cy="32" r="30" fill="#ffc53d" stroke="#1a1526" stroke-width="2.5"/>
          <polygon points="16,24 10,8 26,16" fill="#ff9f1c" stroke="#1a1526" stroke-width="2.5" stroke-linejoin="round"/>
          <polygon points="48,24 54,8 38,16" fill="#ff9f1c" stroke="#1a1526" stroke-width="2.5" stroke-linejoin="round"/>
          <circle cx="23" cy="32" r="3.5" fill="#1a1526"/>
          <circle cx="41" cy="32" r="3.5" fill="#1a1526"/>
          <circle cx="24.5" cy="30.5" r="1.2" fill="#fff"/>
          <circle cx="42.5" cy="30.5" r="1.2" fill="#fff"/>
          <polygon points="32,37 29,34 35,34" fill="#ff6b6b"/>
          <path d="M28 39c2 2 4 2 4 0 0 2 2 2 4 0" fill="none" stroke="#1a1526" stroke-width="2" stroke-linecap="round"/>
          <line x1="14" y1="34" x2="6" y2="33" stroke="#1a1526" stroke-width="1.8" stroke-linecap="round"/>
          <line x1="14" y1="38" x2="7" y2="40" stroke="#1a1526" stroke-width="1.8" stroke-linecap="round"/>
          <line x1="50" y1="34" x2="58" y2="33" stroke="#1a1526" stroke-width="1.8" stroke-linecap="round"/>
          <line x1="50" y1="38" x2="57" y2="40" stroke="#1a1526" stroke-width="1.8" stroke-linecap="round"/>
        </svg>`
    },
    {
      id: 'monster-purple',
      name: 'Monster Gemoy',
      bg: '#e8dbff',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <rect x="8" y="10" width="48" height="46" rx="20" fill="#b388ff" stroke="#1a1526" stroke-width="2.5"/>
          <circle cx="32" cy="7" r="4" fill="#ffc53d" stroke="#1a1526" stroke-width="2"/>
          <line x1="32" y1="10" x2="32" y2="15" stroke="#1a1526" stroke-width="2.5"/>
          <circle cx="32" cy="27" r="11" fill="#fff" stroke="#1a1526" stroke-width="2.5"/>
          <circle cx="33" cy="27" r="5" fill="#6c4cf1"/>
          <circle cx="34.5" cy="25" r="1.5" fill="#fff"/>
          <path d="M22 44c4 4 16 4 20 0" fill="none" stroke="#1a1526" stroke-width="2.5" stroke-linecap="round"/>
          <polygon points="26,44 28,48 30,44" fill="#fff" stroke="#1a1526" stroke-width="1.2"/>
          <polygon points="34,44 36,48 38,44" fill="#fff" stroke="#1a1526" stroke-width="1.2"/>
        </svg>`
    },
    {
      id: 'dino-green',
      name: 'Dino Cilik',
      bg: '#c8f7dc',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <circle cx="32" cy="32" r="30" fill="#2fd08a" stroke="#1a1526" stroke-width="2.5"/>
          <polygon points="12,18 17,9 23,17" fill="#ffc53d" stroke="#1a1526" stroke-width="2"/>
          <polygon points="26,14 32,5 38,14" fill="#ffc53d" stroke="#1a1526" stroke-width="2"/>
          <polygon points="41,17 47,9 52,18" fill="#ffc53d" stroke="#1a1526" stroke-width="2"/>
          <ellipse cx="23" cy="33" rx="4" ry="5" fill="#1a1526"/>
          <ellipse cx="41" cy="33" rx="4" ry="5" fill="#1a1526"/>
          <circle cx="24.5" cy="31" r="1.3" fill="#fff"/>
          <circle cx="42.5" cy="31" r="1.3" fill="#fff"/>
          <ellipse cx="16" cy="39" rx="3" ry="2" fill="#ff9f9f"/>
          <ellipse cx="48" cy="39" rx="3" ry="2" fill="#ff9f9f"/>
          <path d="M26 43c3 3 9 3 12 0" fill="none" stroke="#1a1526" stroke-width="2.5" stroke-linecap="round"/>
        </svg>`
    },
    {
      id: 'boba-sweet',
      name: 'Boba Ceria',
      bg: '#fed6e3',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <circle cx="32" cy="32" r="30" fill="#ff85a1" stroke="#1a1526" stroke-width="2.5"/>
          <line x1="38" y1="2" x2="34" y2="16" stroke="#46c2ff" stroke-width="5" stroke-linecap="round"/>
          <ellipse cx="23" cy="29" rx="3.5" ry="4" fill="#1a1526"/>
          <ellipse cx="41" cy="29" rx="3.5" ry="4" fill="#1a1526"/>
          <circle cx="24" cy="27.5" r="1.2" fill="#fff"/>
          <circle cx="42" cy="27.5" r="1.2" fill="#fff"/>
          <path d="M27 37c3 4 7 4 10 0" fill="none" stroke="#1a1526" stroke-width="2.2" stroke-linecap="round"/>
          <circle cx="20" cy="48" r="3.5" fill="#3b2a00"/>
          <circle cx="32" cy="50" r="3.5" fill="#3b2a00"/>
          <circle cx="44" cy="47" r="3.5" fill="#3b2a00"/>
        </svg>`
    },
    {
      id: 'star-magic',
      name: 'Bintang Ajaib',
      bg: '#fff2a8',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <polygon points="32,4 40,22 60,24 45,38 50,58 32,48 14,58 19,38 4,24 24,22" fill="#ffce5c" stroke="#1a1526" stroke-width="2.5" stroke-linejoin="round"/>
          <circle cx="26" cy="32" r="3" fill="#1a1526"/>
          <circle cx="38" cy="32" r="3" fill="#1a1526"/>
          <circle cx="27" cy="31" r="1" fill="#fff"/>
          <circle cx="39" cy="31" r="1" fill="#fff"/>
          <path d="M28 39c2 2 6 2 8 0" fill="none" stroke="#1a1526" stroke-width="2" stroke-linecap="round"/>
          <circle cx="21" cy="36" r="2.5" fill="#ff8fa3" opacity="0.8"/>
          <circle cx="43" cy="36" r="2.5" fill="#ff8fa3" opacity="0.8"/>
        </svg>`
    },
    {
      id: 'robot-cool',
      name: 'Robot Pensil',
      bg: '#cdeeff',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <rect x="12" y="14" width="40" height="38" rx="8" fill="#46c2ff" stroke="#1a1526" stroke-width="2.5"/>
          <line x1="32" y1="4" x2="32" y2="14" stroke="#1a1526" stroke-width="2.5"/>
          <circle cx="32" cy="4" r="3" fill="#ff6b6b" stroke="#1a1526" stroke-width="1.8"/>
          <rect x="18" y="22" width="10" height="9" rx="3" fill="#fff" stroke="#1a1526" stroke-width="2"/>
          <circle cx="23" cy="26" r="2.2" fill="#1a1526"/>
          <rect x="36" y="22" width="10" height="9" rx="3" fill="#fff" stroke="#1a1526" stroke-width="2"/>
          <circle cx="41" cy="26" r="2.2" fill="#1a1526"/>
          <rect x="22" y="38" width="20" height="6" rx="3" fill="#ffc53d" stroke="#1a1526" stroke-width="2"/>
          <line x1="28" y1="38" x2="28" y2="44" stroke="#1a1526" stroke-width="1.5"/>
          <line x1="36" y1="38" x2="36" y2="44" stroke="#1a1526" stroke-width="1.5"/>
        </svg>`
    },
    {
      id: 'bunny-cloud',
      name: 'Kelinci Gemas',
      bg: '#f1ebff',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <ellipse cx="22" cy="18" rx="6" ry="14" fill="#fff" stroke="#1a1526" stroke-width="2.5"/>
          <ellipse cx="22" cy="18" rx="3" ry="9" fill="#ffb3c6"/>
          <ellipse cx="42" cy="18" rx="6" ry="14" fill="#fff" stroke="#1a1526" stroke-width="2.5"/>
          <ellipse cx="42" cy="18" rx="3" ry="9" fill="#ffb3c6"/>
          <ellipse cx="32" cy="38" rx="24" ry="20" fill="#fff" stroke="#1a1526" stroke-width="2.5"/>
          <circle cx="24" cy="37" r="3" fill="#1a1526"/>
          <circle cx="40" cy="37" r="3" fill="#1a1526"/>
          <circle cx="25" cy="36" r="1" fill="#fff"/>
          <circle cx="41" cy="36" r="1" fill="#fff"/>
          <ellipse cx="32" cy="42" rx="2" ry="1.5" fill="#ff6b6b"/>
          <path d="M29 44c1.5 1.5 3 1.5 3 0 0 1.5 1.5 1.5 3 0" fill="none" stroke="#1a1526" stroke-width="1.8" stroke-linecap="round"/>
        </svg>`
    },
    {
      id: 'fox-ninja',
      name: 'Rubah Cerdik',
      bg: '#ffe0d2',
      svg: `
        <svg viewBox="0 0 64 64" class="avatar-svg" aria-hidden="true">
          <circle cx="32" cy="32" r="30" fill="#ff7a45" stroke="#1a1526" stroke-width="2.5"/>
          <polygon points="12,22 10,6 26,16" fill="#1a1526"/>
          <polygon points="52,22 54,6 38,16" fill="#1a1526"/>
          <path d="M12 32c8 14 20 20 20 20s12-6 20-20c0 0-8-2-20-2s-20 2-20 2Z" fill="#ffffff" stroke="#1a1526" stroke-width="2"/>
          <ellipse cx="22" cy="31" rx="3" ry="2" fill="#1a1526" transform="rotate(-15 22 31)"/>
          <ellipse cx="42" cy="31" rx="3" ry="2" fill="#1a1526" transform="rotate(15 42 31)"/>
          <polygon points="32,48 29,44 35,44" fill="#1a1526"/>
        </svg>`
    }
  ];

  const AvatarRegistry = {
    getAll() {
      return AVATARS;
    },

    list() {
      return AVATARS;
    },

    getById(id) {
      return AVATARS.find(a => a.id === id) || AVATARS[0];
    },

    render(id, customClass = '') {
      const avatar = this.getById(id);
      return `<span class="avatar-wrap ${customClass}" data-avatar-id="${avatar.id}" style="--avatar-bg: ${avatar.bg}" title="${avatar.name}">${avatar.svg}</span>`;
    },

    renderGrid(containerEl, selectedId, onSelect) {
      if (!containerEl) return;
      containerEl.innerHTML = '';

      AVATARS.forEach((avatar) => {
        const label = document.createElement('label');
        label.className = 'avatar-option';
        label.title = avatar.name;

        const input = document.createElement('input');
        input.type = 'radio';
        input.name = 'avatar';
        input.value = avatar.id;
        input.checked = avatar.id === selectedId;

        const preview = document.createElement('div');
        preview.className = 'avatar-preview';
        preview.style.setProperty('--avatar-bg', avatar.bg);
        preview.innerHTML = avatar.svg;

        input.addEventListener('change', () => {
          if (input.checked && typeof onSelect === 'function') {
            onSelect(avatar.id);
          }
        });

        label.appendChild(input);
        label.appendChild(preview);
        containerEl.appendChild(label);
      });
    }
  };

  window.AvatarRegistry = AvatarRegistry;
})(window);

