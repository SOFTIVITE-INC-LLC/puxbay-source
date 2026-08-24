const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'utf8');

const fullscreenProps = `
  isFullscreen = signal(false);

  constructor() {
    window.addEventListener('online', () => this.isOffline.set(false));
    window.addEventListener('offline', () => this.isOffline.set(true));
    
    // Check initial theme
    if (document.documentElement.classList.contains('dark')) {
      this.theme.set('dark');
    }

    document.addEventListener('fullscreenchange', () => {
      this.isFullscreen.set(!!document.fullscreenElement);
    });
  }

  toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(err => {
        this.toastr.error('Error attempting to enable fullscreen mode');
      });
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
  }
`;

code = code.replace(`  constructor() {
    window.addEventListener('online', () => this.isOffline.set(false));
    window.addEventListener('offline', () => this.isOffline.set(true));
    
    // Check initial theme
    if (document.documentElement.classList.contains('dark')) {
      this.theme.set('dark');
    }
  }`, fullscreenProps);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', code);
