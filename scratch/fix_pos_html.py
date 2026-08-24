import re

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'r') as f:
    content = f.read()

# Replace activeModal checks
content = content.replace("facade.activeModal === 'success'", "facade.isSuccessModalOpen()")
content = content.replace("facade.activeModal === 'mobile_money'", "facade.isMobileMoneyModalOpen()")
content = content.replace("facade.activeModal === 'pin'", "facade.isPinModalOpen()")
content = content.replace("facade.activeModal === 'hardware'", "facade.isHardwareModalOpen()")
content = content.replace("facade.activeModal === 'gift_card'", "facade.isGiftCardModalOpen()")
content = content.replace("facade.activeModal === 'batch'", "facade.isBatchModalOpen()")
content = content.replace("facade.activeModal === 'split'", "facade.isSplitModalOpen()")
content = content.replace("facade.activeModal === 'issue_gift_card'", "facade.isIssueGiftCardModalOpen()")

# Add facade. to methods
methods = [
    'closeModal', 'printReceipt', 'clearCart', 'toggleKiosk', 'toggleTheme', 
    'syncInventory', 'openModal', 'handleProductClick', 'updateQuantity', 
    'removeFromCart', 'clearPin', 'handlePinInput', 'deletePin', 'validatePin', 
    'processCheckout', 'setPaymentMethod', 'toggleMobileCart', 'selectCustomer', 'addToCart', 'addGiftCardToCart', 'pairPrinter'
]

for method in methods:
    content = re.sub(r'\(click\)="' + method + r'(\s*[\(\)])', r'(click)="facade.' + method + r'\1', content)
    content = re.sub(r'\(click\)="' + method + r'"', r'(click)="facade.' + method + r'()"', content)
    
# Computed properties missing facade
content = content.replace("cartSubtotal", "facade.cartSubtotal()")
content = content.replace("cartTotal", "facade.cartTotal()")
content = content.replace("splitRemaining", "facade.splitRemaining()")
content = content.replace("filteredProducts", "facade.filteredProducts()")
content = content.replace("selectedCustomer", "facade.selectedCustomer()")
content = content.replace("filteredCustomers", "facade.customers()") # simple fallback

# Replace vue specific stuff
content = content.replace("facade.currency{{", "facade.currency(){{") 
content = content.replace("facade.currency(){{", "facade.currency() }}{{") 

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'w') as f:
    f.write(content)
