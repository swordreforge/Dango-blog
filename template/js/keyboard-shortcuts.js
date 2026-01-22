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

// 管理员面板聚焦模式快捷键管理器
class AdminKeyboardManager {
  constructor() {
    this.isFocusMode = false;
    this.currentTab = 'articles';
    this.activeModal = null;
    this.selectedRows = new Set();
    this.selectedFile = null;
    this.currentPath = '/';
    
    this.tabs = [
      'articles', 'users', 'comments', 'categories', 'tags',
      'analytics', 'about', 'filemanager', 'attachments', 'settings'
    ];
    
    this.init();
  }

  init() {
    // 检查是否在管理员页面
    if (!this.isAdminPage()) {
      return;
    }

    // 监听键盘事件
    document.addEventListener('keydown', this.handleKeyDown.bind(this));
    document.addEventListener('keyup', this.handleKeyUp.bind(this));
    
    // 监听模态框变化
    this.observeModals();
    
    // 监听标签页变化
    this.observeTabs();
    
    console.log('[管理员快捷键] 已初始化');
  }

  isAdminPage() {
    return window.location.pathname === '/admin' || window.location.pathname.startsWith('/admin');
  }

  handleKeyDown(event) {
    // 如果在输入框、文本域或富文本编辑器中，禁用部分快捷键
    if (this.isInputElement(event.target)) {
      return this.handleInputShortcuts(event);
    }

    const key = this.getKeyString(event);

    // 全局快捷键（始终可用）
    if (this.handleGlobalShortcuts(key, event)) {
      return;
    }

    // 聚焦模式检查
    if (!this.isFocusMode) {
      return;
    }

    // 模态框内快捷键
    if (this.activeModal) {
      if (this.handleModalShortcuts(key, event)) {
        return;
      }
    }

    // 当前标签页特定快捷键
    if (this.handleTabShortcuts(key, event)) {
      return;
    }

    // 通用聚焦模式快捷键
    if (this.handleFocusModeShortcuts(key, event)) {
      return;
    }
  }

  handleKeyUp(event) {
    // 处理按键释放事件（如果需要）
  }

  isInputElement(element) {
    return element && (
      element.tagName === 'INPUT' ||
      element.tagName === 'TEXTAREA' ||
      element.tagName === 'SELECT' ||
      element.isContentEditable
    );
  }

  handleInputShortcuts(event) {
    const key = this.getKeyString(event);
    
    // 在输入框中只处理 Escape 键
    if (key === 'Escape') {
      if (this.activeModal) {
        this.closeCurrentModal();
        event.preventDefault();
        return true;
      }
      if (this.isFocusMode) {
        this.exitFocusMode();
        event.preventDefault();
        return true;
      }
    }
    
    return false;
  }

  getKeyString(event) {
    // 构建按键字符串
    let key = event.key;
    
    // 处理特殊键
    if (key === ' ') {
      key = 'Space';
    }
    
    return key;
  }

  handleGlobalShortcuts(key, event) {
    switch(key) {
      case 'i':
        if (!this.activeModal) {
          this.enterFocusMode();
          event.preventDefault();
          return true;
        }
        break;
        
      case 'q':
        this.exitFocusMode();
        event.preventDefault();
        return true;
        
      case 'Escape':
        if (this.activeModal) {
          this.closeCurrentModal();
        } else if (this.isFocusMode) {
          this.exitFocusMode();
        }
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleModalShortcuts(key, event) {
    if (!this.activeModal) return false;
    
    switch(key) {
      case 'Escape':
        this.closeCurrentModal();
        event.preventDefault();
        return true;
        
      case 'Enter':
        // 查找并点击主要操作按钮
        const primaryBtn = this.activeModal.querySelector('.btn-primary');
        if (primaryBtn && !this.isInputElement(document.activeElement)) {
          primaryBtn.click();
          event.preventDefault();
          return true;
        }
        break;
        
      case 's':
        // 查找并点击保存/提交按钮
        const submitBtn = this.activeModal.querySelector('button[type="submit"], .btn-primary');
        if (submitBtn && !this.isInputElement(document.activeElement)) {
          submitBtn.click();
          event.preventDefault();
          return true;
        }
        break;
    }
    return false;
  }

  handleTabShortcuts(key, event) {
    // 标签页切换 (1-0)
    if (key >= '0' && key <= '9') {
      const tabIndex = key === '0' ? 9 : parseInt(key) - 1;
      
      if (tabIndex < this.tabs.length) {
        this.switchToTab(this.tabs[tabIndex]);
        event.preventDefault();
        return true;
      }
    }
    
    // 标签页内导航
    switch(key) {
      case 'ArrowRight':
        this.nextTab();
        event.preventDefault();
        return true;
        
      case 'ArrowLeft':
        this.previousTab();
        event.preventDefault();
        return true;
        
      case 'r':
        this.refreshCurrentTab();
        event.preventDefault();
        return true;
        
      case 'n':
        this.createNewItem();
        event.preventDefault();
        return true;
        
      case 'u':
        this.uploadItem();
        event.preventDefault();
        return true;
        
      case 'f':
        this.openSearch();
        event.preventDefault();
        return true;
    }
    
    // 特定标签页功能
    return this.handleSpecificTabShortcuts(key, event);
  }

  handleSpecificTabShortcuts(key, event) {
    switch(this.currentTab) {
      case 'articles':
        return this.handleArticleShortcuts(key, event);
      case 'filemanager':
        return this.handleFileManagerShortcuts(key, event);
      case 'users':
        return this.handleUserShortcuts(key, event);
      case 'comments':
        return this.handleCommentShortcuts(key, event);
      case 'categories':
        return this.handleCategoryShortcuts(key, event);
      case 'tags':
        return this.handleTagShortcuts(key, event);
      case 'attachments':
        return this.handleAttachmentShortcuts(key, event);
    }
    return false;
  }

  handleArticleShortcuts(key, event) {
    if (!this.selectedRows.size && !this.hasSelectedRow()) return false;
    
    switch(key) {
      case 'e':
        this.editSelectedArticle();
        event.preventDefault();
        return true;
        
      case 'd':
        this.deleteSelectedArticle();
        event.preventDefault();
        return true;
        
      case 'v':
        this.viewSelectedArticle();
        event.preventDefault();
        return true;
        
      case 'a':
        this.attachToSelectedArticle();
        event.preventDefault();
        return true;
        
      case 'p':
        this.publishSelectedArticle();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleFileManagerShortcuts(key, event) {
    switch(key) {
      case 'Enter':
        this.openSelectedFile();
        event.preventDefault();
        return true;
        
      case 'Backspace':
        this.goUpDirectory();
        event.preventDefault();
        return true;
        
      case 'r':
        if (!this.selectedFile) {
          this.refreshCurrentTab();
        } else {
          this.renameSelectedFile();
        }
        event.preventDefault();
        return true;
        
      case 'Delete':
        this.deleteSelectedFile();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleUserShortcuts(key, event) {
    if (!this.selectedRows.size && !this.hasSelectedRow()) return false;
    
    switch(key) {
      case 'e':
        this.editSelectedUser();
        event.preventDefault();
        return true;
        
      case 'd':
        this.deleteSelectedUser();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleCommentShortcuts(key, event) {
    if (!this.selectedRows.size && !this.hasSelectedRow()) return false;
    
    switch(key) {
      case 'a':
        this.approveSelectedComment();
        event.preventDefault();
        return true;
        
      case 'd':
        this.deleteSelectedComment();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleCategoryShortcuts(key, event) {
    if (!this.selectedRows.size && !this.hasSelectedRow()) return false;
    
    switch(key) {
      case 'e':
        this.editSelectedCategory();
        event.preventDefault();
        return true;
        
      case 'd':
        this.deleteSelectedCategory();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleTagShortcuts(key, event) {
    if (!this.selectedRows.size && !this.hasSelectedRow()) return false;
    
    switch(key) {
      case 'e':
        this.editSelectedTag();
        event.preventDefault();
        return true;
        
      case 'd':
        this.deleteSelectedTag();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleAttachmentShortcuts(key, event) {
    if (!this.selectedRows.size && !this.hasSelectedRow()) return false;
    
    switch(key) {
      case 'v':
        this.viewSelectedAttachment();
        event.preventDefault();
        return true;
        
      case 'e':
        this.editSelectedAttachment();
        event.preventDefault();
        return true;
        
      case 'd':
        this.deleteSelectedAttachment();
        event.preventDefault();
        return true;
    }
    return false;
  }

  handleFocusModeShortcuts(key, event) {
    // 通用聚焦模式快捷键
    switch(key) {
      case 'Tab':
        event.preventDefault();
        return true;
        
      case '?':
        this.showAdminShortcutHelp();
        event.preventDefault();
        return true;
    }
    return false;
  }

  // ========== 聚焦模式控制 ==========
  
  enterFocusMode() {
    if (this.isFocusMode) return;
    
    this.isFocusMode = true;
    document.body.classList.add('admin-focus-mode');
    
    // 禁用普通快捷键
    if (window.keyboardShortcuts) {
      window.keyboardShortcuts.disable();
    }
    
    this.showToast('已进入管理员聚焦模式', 'success');
    console.log('[管理员快捷键] 进入聚焦模式');
  }

  exitFocusMode() {
    if (!this.isFocusMode) return;
    
    this.isFocusMode = false;
    document.body.classList.remove('admin-focus-mode');
    
    // 启用普通快捷键
    if (window.keyboardShortcuts) {
      window.keyboardShortcuts.enable();
    }
    
    this.showToast('已退出管理员聚焦模式', 'info');
    console.log('[管理员快捷键] 退出聚焦模式');
  }

  // ========== 标签页操作 ==========
  
  switchToTab(tabId) {
    const tabButton = document.querySelector(`.tab-btn[data-tab="${tabId}"]`);
    if (tabButton) {
      tabButton.click();
      this.currentTab = tabId;
      console.log(`[管理员快捷键] 切换到标签页: ${tabId}`);
    }
  }

  nextTab() {
    const currentIndex = this.tabs.indexOf(this.currentTab);
    const nextIndex = (currentIndex + 1) % this.tabs.length;
    this.switchToTab(this.tabs[nextIndex]);
  }

  previousTab() {
    const currentIndex = this.tabs.indexOf(this.currentTab);
    const prevIndex = (currentIndex - 1 + this.tabs.length) % this.tabs.length;
    this.switchToTab(this.tabs[prevIndex]);
  }

  refreshCurrentTab() {
    // 根据当前标签页查找对应的刷新按钮
    let refreshBtn = null;
    
    // 文章管理 - 使用特定的ID
    if (this.currentTab === 'articles') {
      refreshBtn = document.getElementById('refreshArticlesBtn');
    }
    // 附件管理 - 使用特定的ID
    else if (this.currentTab === 'attachments') {
      refreshBtn = document.getElementById('amRefreshBtn');
    }
    // 文件管理
    else if (this.currentTab === 'filemanager') {
      refreshBtn = document.getElementById('fmRefreshBtn');
    }
    // 通用查找
    else {
      refreshBtn = document.querySelector(`#${this.currentTab}RefreshBtn, .refresh-btn`);
    }
    
    if (refreshBtn) {
      refreshBtn.click();
      this.showToast('已刷新', 'success');
      return;
    }
    
    // 尝试通过文本查找（在当前标签页内容区域内）
    const currentTabPane = document.querySelector(`.tab-pane[data-tab="${this.currentTab}"], .tab-pane.active`);
    if (currentTabPane) {
      const buttons = currentTabPane.querySelectorAll('button');
      for (const btn of buttons) {
        const text = btn.textContent.trim();
        if (text === '刷新' || text.includes('刷新')) {
          btn.click();
          this.showToast('已刷新', 'success');
          return;
        }
      }
    }
    
    // 尝试调用刷新函数（如果存在）
    if (this.currentTab === 'articles' && typeof loadPassages === 'function') {
      loadPassages();
      this.showToast('已刷新', 'success');
    } else if (this.currentTab === 'attachments' && typeof loadAttachments === 'function') {
      loadAttachments();
      this.showToast('已刷新', 'success');
    } else if (this.currentTab === 'filemanager' && window.FileManager) {
      FileManager.loadFiles();
      this.showToast('已刷新', 'success');
    }
  }

  createNewItem() {
    // 根据当前标签页查找对应的新建按钮
    let newBtn = null;
    
    // 文章管理 - 使用特定的ID
    if (this.currentTab === 'articles') {
      newBtn = document.getElementById('newArticleBtn');
    }
    // 用户管理
    else if (this.currentTab === 'users') {
      newBtn = document.getElementById('newUserBtn');
    }
    // 分类管理
    else if (this.currentTab === 'categories') {
      newBtn = document.getElementById('newCategoryBtn');
    }
    // 标签管理
    else if (this.currentTab === 'tags') {
      newBtn = document.getElementById('newTagBtn');
    }
    // 通用查找
    else {
      newBtn = document.querySelector(`#${this.currentTab}NewBtn, .new-btn`);
    }
    
    if (newBtn) {
      newBtn.click();
      return;
    }
    
    // 尝试通过文本查找（在当前标签页内容区域内）
    const currentTabPane = document.querySelector(`.tab-pane[data-tab="${this.currentTab}"], .tab-pane.active`);
    if (currentTabPane) {
      const buttons = currentTabPane.querySelectorAll('button');
      for (const btn of buttons) {
        const text = btn.textContent.trim();
        if (text === '新建文章' || text === '新建用户' || text === '新建分类' || 
            text === '新建标签' || text.startsWith('新建') || text.startsWith('创建') || text.startsWith('添加')) {
          btn.click();
          return;
        }
      }
    }
    
    // 最后尝试在整个页面查找
    const allButtons = document.querySelectorAll('button');
    for (const btn of allButtons) {
      const text = btn.textContent.trim();
      if (text === '新建文章' || text === '新建用户' || text === '新建分类' || 
          text === '新建标签' || text.startsWith('新建')) {
        btn.click();
        return;
      }
    }
    
    this.showToast('未找到新建按钮', 'warning');
  }

  uploadItem() {
    // 根据当前标签页查找对应的上传按钮
    let uploadBtn = null;
    
    // 附件管理 - 使用特定的ID
    if (this.currentTab === 'attachments') {
      uploadBtn = document.getElementById('amUploadBtn');
    }
    // 文件管理
    else if (this.currentTab === 'filemanager') {
      uploadBtn = document.getElementById('fmUploadBtn');
    }
    // 通用查找
    else {
      uploadBtn = document.querySelector(`#${this.currentTab}UploadBtn, .upload-btn`);
    }
    
    if (uploadBtn) {
      uploadBtn.click();
      return;
    }
    
    // 尝试通过文本查找（在当前标签页内容区域内）
    const currentTabPane = document.querySelector(`.tab-pane[data-tab="${this.currentTab}"], .tab-pane.active`);
    if (currentTabPane) {
      const buttons = currentTabPane.querySelectorAll('button');
      for (const btn of buttons) {
        const text = btn.textContent.trim();
        if (text.includes('上传')) {
          btn.click();
          return;
        }
      }
    }
    
    // 最后尝试在整个页面查找
    const allButtons = document.querySelectorAll('button');
    for (const btn of allButtons) {
      const text = btn.textContent.trim();
      if (text.includes('上传')) {
        btn.click();
        return;
      }
    }
    
    this.showToast('未找到上传按钮', 'warning');
  }

  openSearch() {
    // 根据当前标签页查找对应的搜索框
    let searchInput = null;
    
    // 附件管理 - 使用特定的ID
    if (this.currentTab === 'attachments') {
      searchInput = document.getElementById('amSearchInput');
    }
    // 文件管理
    else if (this.currentTab === 'filemanager') {
      searchInput = document.getElementById('fmSearchInput');
    }
    // 文章管理
    else if (this.currentTab === 'articles') {
      searchInput = document.getElementById('articlesSearchInput');
    }
    // 通用查找
    else {
      searchInput = document.querySelector(`#${this.currentTab}SearchInput, .search-input, input[type="search"]`);
    }
    
    if (searchInput) {
      searchInput.focus();
      this.showToast('已聚焦到搜索框', 'success');
      return;
    }
    
    // 尝试在当前标签页内容区域内查找
    const currentTabPane = document.querySelector(`.tab-pane[data-tab="${this.currentTab}"], .tab-pane.active`);
    if (currentTabPane) {
      const inputs = currentTabPane.querySelectorAll('input[type="text"], input[type="search"]');
      for (const input of inputs) {
        const placeholder = input.placeholder || '';
        if (placeholder.includes('搜索') || placeholder.includes('筛选') || placeholder.includes('查找')) {
          input.focus();
          this.showToast('已聚焦到搜索框', 'success');
          return;
        }
      }
    }
    
    this.showToast('未找到搜索框', 'warning');
  }

  // ========== 文章管理操作 ==========
  
  hasSelectedRow() {
    return document.querySelector('.data-table tr.selected') !== null;
  }

  getSelectedRow() {
    return document.querySelector('.data-table tr.selected');
  }

  editSelectedArticle() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const editBtn = selectedRow.querySelector('.btn-edit');
      if (editBtn) {
        editBtn.click();
      }
    }
  }

  deleteSelectedArticle() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const deleteBtn = selectedRow.querySelector('.btn-delete');
      if (deleteBtn) {
        deleteBtn.click();
      }
    }
  }

  viewSelectedArticle() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const viewBtn = selectedRow.querySelector('.btn-view');
      if (viewBtn) {
        viewBtn.click();
      }
    }
  }

  attachToSelectedArticle() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const uploadBtn = selectedRow.querySelector('.btn-upload');
      if (uploadBtn) {
        uploadBtn.click();
      }
    }
  }

  publishSelectedArticle() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      // 查找发布按钮
      const publishBtn = selectedRow.querySelector('.btn-publish, .btn-primary:contains("发布")');
      if (publishBtn) {
        publishBtn.click();
      }
    }
  }

  // ========== 文件管理操作 ==========
  
  openSelectedFile() {
    if (this.selectedFile) {
      if (window.FileManager) {
        window.FileManager.openFile(this.selectedFile.path);
      }
    }
  }

  goUpDirectory() {
    if (window.FileManager) {
      window.FileManager.goBack();
    }
  }

  renameSelectedFile() {
    if (this.selectedFile) {
      if (window.FileManager) {
        window.FileManager.openRenameModal();
      }
    }
  }

  deleteSelectedFile() {
    if (this.selectedFile) {
      if (window.FileManager) {
        window.FileManager.openDeleteModal();
      }
    }
  }

  // ========== 用户管理操作 ==========
  
  editSelectedUser() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const editBtn = selectedRow.querySelector('.btn-edit');
      if (editBtn) {
        editBtn.click();
      }
    }
  }

  deleteSelectedUser() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const deleteBtn = selectedRow.querySelector('.btn-delete');
      if (deleteBtn) {
        deleteBtn.click();
      }
    }
  }

  // ========== 评论管理操作 ==========
  
  approveSelectedComment() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const approveBtn = selectedRow.querySelector('.btn-approve, .btn-primary:contains("批准")');
      if (approveBtn) {
        approveBtn.click();
      }
    }
  }

  deleteSelectedComment() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const deleteBtn = selectedRow.querySelector('.btn-delete');
      if (deleteBtn) {
        deleteBtn.click();
      }
    }
  }

  // ========== 分类管理操作 ==========
  
  editSelectedCategory() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const editBtn = selectedRow.querySelector('.btn-edit');
      if (editBtn) {
        editBtn.click();
      }
    }
  }

  deleteSelectedCategory() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const deleteBtn = selectedRow.querySelector('.btn-delete');
      if (deleteBtn) {
        deleteBtn.click();
      }
    }
  }

  // ========== 标签管理操作 ==========
  
  editSelectedTag() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const editBtn = selectedRow.querySelector('.btn-edit');
      if (editBtn) {
        editBtn.click();
      }
    }
  }

  deleteSelectedTag() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const deleteBtn = selectedRow.querySelector('.btn-delete');
      if (deleteBtn) {
        deleteBtn.click();
      }
    }
  }

  // ========== 附件管理操作 ==========
  
  viewSelectedAttachment() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const viewBtn = selectedRow.querySelector('.btn-view');
      if (viewBtn) {
        viewBtn.click();
      }
    }
  }

  editSelectedAttachment() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const editBtn = selectedRow.querySelector('.btn-edit');
      if (editBtn) {
        editBtn.click();
      }
    }
  }

  deleteSelectedAttachment() {
    const selectedRow = this.getSelectedRow();
    if (selectedRow) {
      const deleteBtn = selectedRow.querySelector('.btn-delete');
      if (deleteBtn) {
        deleteBtn.click();
      }
    }
  }

  // ========== 模态框操作 ==========
  
  observeModals() {
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        mutation.addedNodes.forEach((node) => {
          if (node.classList && node.classList.contains('modal') && node.classList.contains('active')) {
            this.activeModal = node;
          }
        });
        
        mutation.removedNodes.forEach((node) => {
          if (node.classList && node.classList.contains('modal')) {
            this.activeModal = null;
          }
        });
      });
    });
    
    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }

  closeCurrentModal() {
    if (this.activeModal) {
      const closeBtn = this.activeModal.querySelector('.modal-close');
      if (closeBtn) {
        closeBtn.click();
      } else {
        this.activeModal.classList.remove('active');
      }
    }
  }

  // ========== 标签页监听 ==========
  
  observeTabs() {
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
          const target = mutation.target;
          if (target.classList && target.classList.contains('tab-btn')) {
            if (target.classList.contains('active')) {
              this.currentTab = target.dataset.tab;
              console.log(`[管理员快捷键] 当前标签页: ${this.currentTab}`);
            }
          }
        }
      });
    });
    
    document.querySelectorAll('.tab-btn').forEach(tab => {
      observer.observe(tab, { attributes: true });
    });
  }

  // ========== 帮助界面 ==========
  
  showAdminShortcutHelp() {
    const helpContent = `
      <div style="padding: 20px; max-width: 600px;">
        <h3 style="margin-bottom: 15px; color: rgba(255,255,255,0.9);">管理员聚焦模式快捷键</h3>
        
        <h4 style="color: rgba(255,183,122,0.9); margin-top: 20px;">聚焦模式控制</h4>
        <ul style="color: rgba(255,255,255,0.7); line-height: 1.8;">
          <li><kbd style="background: rgba(255,183,122,0.2); padding: 2px 8px; border-radius: 4px; border: 1px solid rgba(255,183,122,0.5);">i</kbd> - 进入聚焦模式</li>
          <li><kbd style="background: rgba(255,183,122,0.2); padding: 2px 8px; border-radius: 4px; border: 1px solid rgba(255,183,122,0.5);">q</kbd> - 退出聚焦模式</li>
          <li><kbd style="background: rgba(255,183,122,0.2); padding: 2px 8px; border-radius: 4px; border: 1px solid rgba(255,183,122,0.5);">Esc</kbd> - 退出聚焦模式/关闭模态框</li>
        </ul>
        
        <h4 style="color: rgba(255,183,122,0.9); margin-top: 20px;">标签页切换</h4>
        <ul style="color: rgba(255,255,255,0.7); line-height: 1.8;">
          <li><kbd>1</kbd> - 文章管理</li>
          <li><kbd>2</kbd> - 用户管理</li>
          <li><kbd>3</kbd> - 评论管理</li>
          <li><kbd>4</kbd> - 分类管理</li>
          <li><kbd>5</kbd> - 标签管理</li>
          <li><kbd>6</kbd> - 统计分析</li>
          <li><kbd>7</kbd> - 关于页面</li>
          <li><kbd>8</kbd> - 文件管理</li>
          <li><kbd>9</kbd> - 附件管理</li>
          <li><kbd>0</kbd> - 系统设置</li>
          <li><kbd>← →</kbd> - 切换标签页</li>
        </ul>
        
        <h4 style="color: rgba(255,183,122,0.9); margin-top: 20px;">通用操作</h4>
        <ul style="color: rgba(255,255,255,0.7); line-height: 1.8;">
          <li><kbd>r</kbd> - 刷新当前数据</li>
          <li><kbd>n</kbd> - 新建项目</li>
          <li><kbd>u</kbd> - 上传</li>
          <li><kbd>f</kbd> - 搜索/筛选</li>
        </ul>
        
        <h4 style="color: rgba(255,183,122,0.9); margin-top: 20px;">文章管理</h4>
        <ul style="color: rgba(255,255,255,0.7); line-height: 1.8;">
          <li><kbd>e</kbd> - 编辑选中文章</li>
          <li><kbd>d</kbd> - 删除选中文章</li>
          <li><kbd>v</kbd> - 查看详情</li>
          <li><kbd>a</kbd> - 上传附件</li>
          <li><kbd>p</kbd> - 发布文章</li>
        </ul>
        
        <h4 style="color: rgba(255,183,122,0.9); margin-top: 20px;">文件管理</h4>
        <ul style="color: rgba(255,255,255,0.7); line-height: 1.8;">
          <li><kbd>Enter</kbd> - 打开选中项</li>
          <li><kbd>Backspace</kbd> - 返回上级目录</li>
          <li><kbd>r</kbd> - 重命名</li>
          <li><kbd>Delete</kbd> - 删除</li>
        </ul>
        
        <h4 style="color: rgba(255,183,122,0.9); margin-top: 20px;">模态框操作</h4>
        <ul style="color: rgba(255,255,255,0.7); line-height: 1.8;">
          <li><kbd>Esc</kbd> - 关闭模态框</li>
          <li><kbd>Enter</kbd> - 确认/主要操作</li>
          <li><kbd>s</kbd> - 保存/提交</li>
        </ul>
      </div>
    `;
    
    // 创建帮助模态框
    const modal = document.createElement('div');
    modal.className = 'modal active';
    modal.innerHTML = `
      <div class="modal-content" style="max-width: 650px;">
        <div class="modal-header">
          <h3>管理员聚焦模式快捷键</h3>
          <button class="modal-close" onclick="this.closest('.modal').remove()">×</button>
        </div>
        <div class="modal-body">
          ${helpContent}
        </div>
      </div>
    `;
    
    document.body.appendChild(modal);
  }

  // ========== 工具方法 ==========
  
  showToast(message, type = 'info') {
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
      const container = document.createElement('div');
      container.id = 'toastContainer';
      container.className = 'toast-container';
      document.body.appendChild(container);
      container.appendChild(toast);
    }

    setTimeout(() => {
      toast.classList.add('closing');
      setTimeout(() => toast.remove(), 300);
    }, 2000);

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
}

// 页面加载完成后初始化快捷键系统
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    window.keyboardShortcuts = new KeyboardShortcuts();
    // 初始化管理员快捷键系统
    window.adminKeyboardManager = new AdminKeyboardManager();
  });
} else {
  window.keyboardShortcuts = new KeyboardShortcuts();
  // 初始化管理员快捷键系统
  window.adminKeyboardManager = new AdminKeyboardManager();
}