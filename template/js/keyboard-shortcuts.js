// 键盘快捷键系统
// 为导航栏按钮设置快捷键

class KeyboardShortcuts {
  constructor() {
    this.shortcuts = {
      // 导航快捷键
      '1': { action: 'navigate', url: '/', label: '主页' },
      '2': { action: 'navigate', url: '/passage', label: '文章' },
      '3': { action: 'navigate', url: '/collect', label: '归档' },
      '4': { action: 'navigate', url: '/about', label: '关于' },
      '5': { action: 'openModal', modalId: 'userCenterModal', label: '个人中心' },
      '6': { action: 'navigate', url: '/markdown-editor', label: '编辑器' },

      // 功能快捷键
      'l': { action: 'openModal', modalId: 'loginModal', label: '登录' },
      'Escape': { action: 'closeAllModals', label: '关闭模态框' },

      // 音乐播放器快捷键
      ' ': { action: 'music', musicAction: 'togglePlay', label: '播放/暂停' },
      'ArrowLeft': { action: 'music', musicAction: 'previous', label: '上一首' },
      'ArrowRight': { action: 'music', musicAction: 'next', label: '下一首' },
      'ArrowUp': { action: 'music', musicAction: 'volumeUp', label: '音量+' },
      'ArrowDown': { action: 'music', musicAction: 'volumeDown', label: '音量-' },
      'm': { action: 'music', musicAction: 'mute', label: '静音' },
      'p': { action: 'music', musicAction: 'playlist', label: '播放列表' },

      // 管理员快捷键
      'a': { action: 'navigate', url: '/admin', label: '管理员设置', adminOnly: true }
    };

    this.enabled = true;
    this.init();
  }

  init() {
    // 监听键盘事件
    document.addEventListener('keydown', (e) => this.handleKeyPress(e));

    // 显示快捷键提示
    this.showShortcutHints();

    // 添加快捷键帮助按钮
    this.addHelpButton();
  }

  handleKeyPress(e) {
    // 如果快捷键功能被禁用,不处理
    if (!this.enabled) return;

    // 如果用户正在输入框中输入,不触发快捷键
    const activeElement = document.activeElement;
    if (activeElement && (
      activeElement.tagName === 'INPUT' ||
      activeElement.tagName === 'TEXTAREA' ||
      activeElement.isContentEditable
    )) {
      return;
    }

    // 检查播放列表是否打开
    const playlist = document.getElementById('musicPlaylist');
    const isPlaylistOpen = playlist && playlist.classList.contains('show');

    // 如果播放列表打开，上下键不触发全局快捷键（由播放列表内部处理）
    if (isPlaylistOpen && (e.key === 'ArrowUp' || e.key === 'ArrowDown')) {
      return;
    }

    // 检查是否在文章页面且处于聚焦模式
    const isPassageFocusMode = document.body.classList.contains('focus-mode');
    if (isPassageFocusMode && (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
      return;
    }

    // 检查是否在归档页面且处于聚焦模式
    const isCollectFocusMode = document.body.classList.contains('collect-focus-mode');
    if (isCollectFocusMode && (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
      return;
    }

    const key = e.key;
    let shortcut = null;

    // 首先尝试使用 e.key 查找快捷键
    if (this.shortcuts[key]) {
      shortcut = this.shortcuts[key];
    } else {
      // 如果 e.key 没有找到，尝试使用 keyCode 映射
      const keyCodeMap = {
        49: '1',  // 数字键1
        50: '2',  // 数字键2
        51: '3',  // 数字键3
        52: '4',  // 数字键4
        53: '5',  // 数字键5
        54: '6',  // 数字键6
        76: 'l',  // 字母键L
      };

      const mappedKey = keyCodeMap[e.keyCode];
      if (mappedKey && this.shortcuts[mappedKey]) {
        shortcut = this.shortcuts[mappedKey];
      }
    }

    // 检查是否有对应的快捷键
    if (shortcut) {
      e.preventDefault();
      e.stopPropagation();

      // 检查是否是管理员专用快捷键
      if (shortcut.adminOnly && !this.isAdmin()) {
        this.showToast('此快捷键仅管理员可用', 'warning');
        return;
      }

      // 执行对应的操作
      this.executeAction(shortcut);
    }
  }

  executeAction(shortcut) {
    switch (shortcut.action) {
      case 'navigate':
        if (window.location.pathname !== shortcut.url) {
          window.location.href = shortcut.url;
        }
        break;

      case 'openModal':
        const modal = document.getElementById(shortcut.modalId);
        if (modal) {
          modal.classList.add('active');
          this.showToast(`已打开: ${shortcut.label}`, 'success');
        }
        break;

      case 'closeAllModals':
        document.querySelectorAll('.modal.active').forEach(modal => {
          modal.classList.remove('active');
        });
        this.showToast('已关闭所有模态框', 'success');
        break;

      case 'music':
        this.executeMusicAction(shortcut.musicAction, shortcut.label);
        break;
    }
  }

  executeMusicAction(musicAction, label) {
    // 检查音乐播放器是否存在且已启用
    if (!window.musicPlayer || !window.musicPlayer.settings || !window.musicPlayer.settings.enabled) {
      this.showToast('音乐播放器未启用', 'warning');
      return;
    }

    const player = window.musicPlayer;

    try {
      switch (musicAction) {
        case 'togglePlay':
          player.togglePlay();
          this.showToast(player.isPlaying ? '正在播放' : '已暂停', 'success');
          break;

        case 'previous':
          player.playPrevious();
          this.showToast('上一首', 'success');
          break;

        case 'next':
          player.playNext();
          this.showToast('下一首', 'success');
          break;

        case 'volumeUp':
          if (player.audio) {
            const newVolume = Math.min(100, player.audio.volume * 100 + 10);
            player.audio.volume = newVolume / 100;
            const volumeBar = document.querySelector('#volumeBar');
            if (volumeBar) {
              volumeBar.value = newVolume;
            }
            player.saveState();
            this.showToast(`音量: ${Math.round(newVolume)}%`, 'success');
          }
          break;

        case 'volumeDown':
          if (player.audio) {
            const newVolume = Math.max(0, player.audio.volume * 100 - 10);
            player.audio.volume = newVolume / 100;
            const volumeBar = document.querySelector('#volumeBar');
            if (volumeBar) {
              volumeBar.value = newVolume;
            }
            player.saveState();
            this.showToast(`音量: ${Math.round(newVolume)}%`, 'success');
          }
          break;

        case 'mute':
          player.toggleMute();
          this.showToast(player.audio.muted ? '已静音' : '已取消静音', 'success');
          break;

        case 'playlist':
          player.togglePlaylist();
          this.showToast('播放列表', 'success');
          break;
      }
    } catch (error) {
      console.error('[音乐播放器快捷键错误]', error);
      this.showToast(`操作失败: ${error.message}`, 'error');
    }
  }

  isAdmin() {
    // 检查是否有管理员元素可见
    const adminElements = document.querySelectorAll('.admin-only');
    return Array.from(adminElements).some(el => {
      return window.getComputedStyle(el).display !== 'none';
    });
  }

  isPassagePage() {
    // 检查是否在文章页面
    return window.location.pathname === '/passage' || window.location.pathname.startsWith('/passage/');
  }

  isCollectPage() {
    // 检查是否在归档页面
    return window.location.pathname === '/collect' || window.location.pathname.startsWith('/collect/');
  }

  showShortcutHints() {
    // 为导航链接添加快捷键提示
    const navLinks = document.querySelectorAll('nav a, nav button');
    
    navLinks.forEach(link => {
      const href = link.getAttribute('href');
      const id = link.getAttribute('id');
      
      let shortcutKey = null;
      let label = null;

      // 根据链接或ID查找对应的快捷键
      Object.entries(this.shortcuts).forEach(([key, shortcut]) => {
        if (shortcut.action === 'navigate' && shortcut.url === href) {
          shortcutKey = key;
          label = shortcut.label;
        } else if (shortcut.action === 'openModal' && shortcut.modalId === id) {
          shortcutKey = key;
          label = shortcut.label;
        }
      });

      // 如果找到快捷键,添加提示
      if (shortcutKey && label) {
        // 检查是否已经存在快捷键提示
        let hint = link.querySelector('.shortcut-hint');
        if (!hint) {
          hint = document.createElement('span');
          hint.className = 'shortcut-hint';
          hint.textContent = shortcutKey;
          link.appendChild(hint);
        }
      }
    });
  }

  addHelpButton() {
    // 添加快捷键帮助按钮
    const helpButton = document.createElement('button');
    helpButton.className = 'shortcuts-help-btn';
    helpButton.innerHTML = `
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
        <line x1="12" y1="17" x2="12.01" y2="17"></line>
      </svg>
      快捷键
    `;
    
    helpButton.addEventListener('click', () => this.showHelpModal());
    
    // 添加到导航栏
    const nav = document.querySelector('nav');
    if (nav) {
      nav.appendChild(helpButton);
    }
  }

  showHelpModal() {
    // 创建帮助模态框
    const helpModal = document.createElement('div');
    helpModal.className = 'modal shortcuts-help-modal active';
    helpModal.innerHTML = `
      <div class="modal-content">
        <div class="modal-header">
          <h3>键盘快捷键</h3>
          <button class="modal-close">&times;</button>
        </div>
        <div class="modal-body">
          <div class="shortcuts-list">
            <h4>导航快捷键</h4>
            ${this.renderShortcutList(['1', '2', '3', '4', '6'])}

            <h4>功能快捷键</h4>
            ${this.renderShortcutList(['5', 'l', 'Escape'])}

            ${this.isPassagePage() ? `
            <h4>文章页面 - 文本聚焦模式</h4>
            <div class="shortcut-item">
              <kbd class="shortcut-key">i</kbd>
              <span class="shortcut-label">进入文本聚焦模式</span>
            </div>
            <div class="shortcut-item">
              <kbd class="shortcut-key">q</kbd>
              <span class="shortcut-label">退出文本聚焦模式</span>
            </div>
            <div class="shortcut-item">
              <kbd class="shortcut-key">ESC</kbd>
              <span class="shortcut-label">暂时退出聚焦模式（可关闭模态框）</span>
            </div>
            <div class="shortcut-description">
              聚焦模式下：← → 切换面板，↑ ↓ 导航，Enter 激活，u 展开/折叠
            </div>
            ` : ''}

            ${this.isCollectPage() ? `
            <h4>归档页面 - 聚焦模式</h4>
            <div class="shortcut-item">
              <kbd class="shortcut-key">i</kbd>
              <span class="shortcut-label">进入聚焦模式</span>
            </div>
            <div class="shortcut-item">
              <kbd class="shortcut-key">q</kbd>
              <span class="shortcut-label">退出聚焦模式</span>
            </div>
            <div class="shortcut-item">
              <kbd class="shortcut-key">ESC</kbd>
              <span class="shortcut-label">返回上一级或暂时退出</span>
            </div>
            <div class="shortcut-description">
              聚焦模式下：↑ ↓ ← → 导航，Enter 进入子菜单/激活，ESC 返回
            </div>
            ` : ''}

            <h4>音乐播放器快捷键</h4>
            ${this.renderShortcutList([' ', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'm', 'p'])}

            ${this.isAdmin() ? '<h4>管理员快捷键</h4>' : ''}
            ${this.isAdmin() ? this.renderShortcutList(['a']) : ''}
          </div>
        </div>
      </div>
    `;

    document.body.appendChild(helpModal);

    // 绑定关闭事件
    const closeBtn = helpModal.querySelector('.modal-close');
    closeBtn.addEventListener('click', () => {
      helpModal.classList.remove('active');
      setTimeout(() => helpModal.remove(), 300);
    });

    // 点击外部关闭
    helpModal.addEventListener('click', (e) => {
      if (e.target === helpModal) {
        helpModal.classList.remove('active');
        setTimeout(() => helpModal.remove(), 300);
      }
    });

    // ESC 键关闭
    const handleEscape = (e) => {
      if (e.key === 'Escape') {
        helpModal.classList.remove('active');
        setTimeout(() => helpModal.remove(), 300);
        document.removeEventListener('keydown', handleEscape);
      }
    };
    document.addEventListener('keydown', handleEscape);
  }

  renderShortcutList(keys) {
    return keys.map(key => {
      const shortcut = this.shortcuts[key];
      if (!shortcut) return '';

      // 格式化按键显示
      let displayKey = key;
      if (key === ' ') {
        displayKey = 'Space';
      } else if (key === 'ArrowLeft') {
        displayKey = '←';
      } else if (key === 'ArrowRight') {
        displayKey = '→';
      } else if (key === 'ArrowUp') {
        displayKey = '↑';
      } else if (key === 'ArrowDown') {
        displayKey = '↓';
      }

      return `
        <div class="shortcut-item">
          <kbd class="shortcut-key">${displayKey}</kbd>
          <span class="shortcut-label">${shortcut.label}</span>
        </div>
      `;
    }).join('');
  }

  showToast(message, type = 'info') {
    // 显示提示消息
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
      <span class="toast-icon">${this.getToastIcon(type)}</span>
      <span class="toast-message">${message}</span>
      <button class="toast-close">&times;</button>
    `;

    const toastContainer = document.getElementById('toastContainer');
    if (toastContainer) {
      toastContainer.appendChild(toast);
    } else {
      // 如果没有 toast 容器,创建一个
      const container = document.createElement('div');
      container.id = 'toastContainer';
      container.className = 'toast-container';
      document.body.appendChild(container);
      container.appendChild(toast);
    }

    // 自动关闭
    setTimeout(() => {
      toast.classList.add('closing');
      setTimeout(() => toast.remove(), 300);
    }, 2000);

    // 手动关闭
    const closeBtn = toast.querySelector('.toast-close');
    closeBtn.addEventListener('click', () => {
      toast.classList.add('closing');
      setTimeout(() => toast.remove(), 300);
    });
  }

  getToastIcon(type) {
    const icons = {
      success: '✓',
      error: '✕',
      warning: '⚠',
      info: 'ℹ'
    };
    return icons[type] || icons.info;
  }

  enable() {
    this.enabled = true;
  }

  disable() {
    this.enabled = false;
  }
}

// 页面加载完成后初始化快捷键系统
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    window.keyboardShortcuts = new KeyboardShortcuts();
  });
} else {
  window.keyboardShortcuts = new KeyboardShortcuts();
}