const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', 'utf8');

if (!code.includes('styles: [')) {
    code = code.replace("  `\n})", "  `,\n  styles: [`\n    @media print {\n      :host {\n        display: block !important;\n      }\n      ::ng-deep body * {\n        visibility: hidden;\n      }\n      ::ng-deep app-receipt, ::ng-deep app-receipt * {\n        visibility: visible;\n      }\n      ::ng-deep app-receipt {\n        position: absolute;\n        left: 0;\n        top: 0;\n        width: 100%;\n        margin: 0;\n        padding: 0;\n      }\n      #print-receipt {\n        display: block !important;\n      }\n    }\n  `]\n})");
    fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', code);
}
