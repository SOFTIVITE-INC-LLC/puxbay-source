from django.core.management.base import BaseCommand
from documentation.models import DocumentationSection, DocumentationArticle
from django.utils.text import slugify

class Command(BaseCommand):
    help = 'Seeds the database with an exhaustive, in-depth API reference with total SDK usage'

    def get_sdk_tabs(self, endpoint_id, **kwargs):
        tabs_config = [
            ('js', 'JavaScript/TS', 'js_code'),
            ('python', 'Python', 'py_code'),
            ('php', 'PHP', 'php_code'),
            ('go', 'Go', 'go_code'),
            ('java', 'Java', 'java_code'),
            ('ruby', 'Ruby', 'ruby_code'),
            ('rust', 'Rust', 'rust_code'),
            ('swift', 'Swift', 'swift_code'),
            ('dart', 'Dart', 'dart_code'),
            ('kotlin', 'Kotlin', 'kt_code'),
            ('dotnet', '.NET', 'dotnet_code')
        ]
        
        btns = []
        contents = []
        count = 0
        for key, label, kwarg_key in tabs_config:
            code = kwargs.get(kwarg_key)
            if code:
                active_class = "active" if count == 0 else ""
                hidden_class = "" if count == 0 else "hidden"
                btns.append(f'<button onclick="switchSDK(\'{key}\', this)" class="sdk-tab-btn {active_class} px-4 py-1.5 text-[10px] font-bold uppercase tracking-widest rounded-lg transition-all">{label}</button>')
                contents.append(f'<div class="sdk-content {key} {hidden_class}"><pre>{code}</pre></div>')
                count += 1

        if not btns: return ""

        return f"""
        <div class="sdk-container my-12" data-endpoint-id="{endpoint_id}">
            <div class="flex flex-wrap items-center gap-1 bg-slate-50 p-1.5 rounded-xl w-full mb-6 border border-slate-200 overflow-x-auto custom-scrollbar shadow-sm">
                {''.join(btns)}
            </div>
            {''.join(contents)}
        </div>
        """

    def handle(self, *args, **options):
        DocumentationArticle.objects.all().delete()
        DocumentationSection.objects.all().delete()

        # Sections
        sections_data = [
            # API Reference Sections
            {'title': 'Authentication & Security', 'order': 1, 'icon': 'lock', 'doc_type': 'api'},
            {'title': 'Product & Inventory', 'order': 2, 'icon': 'inventory_2', 'doc_type': 'api'},
            {'title': 'Customer & CRM', 'order': 3, 'icon': 'group', 'doc_type': 'api'},
            {'title': 'Point of Sale (POS)', 'order': 4, 'icon': 'point_of_sale', 'doc_type': 'api'},
            {'title': 'Supply Chain & Logistics', 'order': 5, 'icon': 'local_shipping', 'doc_type': 'api'},
            {'title': 'Analytics & Intelligence', 'order': 6, 'icon': 'insights', 'doc_type': 'api'},
            {'title': 'Infrastructure & Sync', 'order': 7, 'icon': 'settings_input_component', 'doc_type': 'api'},
            
            # User Manual Sections
            {'title': 'Getting Started', 'order': 1, 'icon': 'rocket_launch', 'doc_type': 'manual'},
            {'title': 'Daily Operations', 'order': 2, 'icon': 'today', 'doc_type': 'manual'},
            {'title': 'Manager Tools', 'order': 3, 'icon': 'admin_panel_settings', 'doc_type': 'manual'},
            {'title': 'Inventory Management', 'order': 4, 'icon': 'inventory', 'doc_type': 'manual'},
        ]
        
        sections = {(s['title'], s['doc_type']): DocumentationSection.objects.create(**s) for s in sections_data}

        # --- MANUAL: GETTING STARTED ---
        DocumentationArticle.objects.create(
            section=sections[('Getting Started', 'manual')],
            title='Opening the Branch',
            slug='manual-opening',
            content="""
                <p class='text-lg mb-6'>Welcome to your first shift! Follow these steps to correctly open your branch in the Puxbay POS system.</p>
                <ol class='list-decimal pl-6 space-y-4'>
                    <li><b>Log In</b>: Enter your staff credentials and verify your identity via the security prompt.</li>
                    <li><b>Register Verification</b>: Count your opening float and enter the amount in the 'Open Register' screen.</li>
                    <li><b>Printer Check</b>: Run a test print to ensure the receipt printer is loaded and online.</li>
                    <li><b>Stock Check</b>: Quickly scan the 'Critical Restock' dashboard for any high-priority items.</li>
                </ol>
            """,
            order=1
        )

        # --- MANUAL: DAILY OPERATIONS ---
        DocumentationArticle.objects.create(
            section=sections[('Daily Operations', 'manual')],
            title='Processing a Sale',
            slug='manual-sale',
            content="""
                <p class='text-lg mb-6'>The core of your work involves serving customers. Here's how to process a standard transaction.</p>
                <ul class='list-disc pl-6 space-y-4'>
                    <li><b>Scanning</b>: Use the barcode scanner or the search bar to add items to the cart.</li>
                    <li><b>Loyalty</b>: Ask the customer if they have a Puxbay Loyalty account. Search by phone number if needed.</li>
                    <li><b>Payment</b>: Select the payment method (Cash, Card, or Wallet) and follow the on-screen prompts.</li>
                    <li><b>Receipt</b>: Once completed, the receipt will print automatically, or you can email it to the customer.</li>
                </ul>
            """,
            order=1
        )

        # --- MANUAL: MANAGER TOOLS ---
        DocumentationArticle.objects.create(
            section=sections[('Manager Tools', 'manual')],
            title='Approving Overrides',
            slug='manual-overrides',
            content="""
                <p class='text-lg mb-6'>Certain actions require manager approval (Staff PIN). Use these for security and audit trail integrity.</p>
                <div class='bg-amber-50 p-6 rounded-xl border border-amber-200 mb-8'>
                    <h4 class='text-amber-900 font-bold mb-2'>Actions Requiring PIN:</h4>
                    <ul class='list-disc pl-6 text-amber-800'>
                        <li>Applying a manual discount greater than 20%.</li>
                        <li>Deleting a completed transaction (Refund).</li>
                        <li>Opening the cash drawer without a sale.</li>
                        <li>Performing a stock adjustment for high-value items.</li>
                    </ul>
                </div>
            """,
            order=1
        )

        # --- 1. AUTHENTICATION ---
        DocumentationArticle.objects.create(
            section=sections[('Authentication & Security', 'api')],
            title='JWT Authentication Flow',
            slug='auth-jwt',
            content=f"""
                <div class='mb-10'>
                    <p class='text-lg'>Puxbay uses <b>Standard JWT (JSON Web Tokens)</b> for secure session management. This flow involves exchanging credentials for an <code>access</code> and <code>refresh</code> token pair.</p>
                </div>
                
                <h4 class='text-slate-900 font-bold mb-4'>Detailed Parameters</h4>
                <table class='w-full text-sm mb-10 border-collapse border border-slate-100 rounded-xl overflow-hidden'>
                    <tr class='bg-slate-50'><th class='p-3 text-left border'>Field</th><th class='p-3 text-left border'>Type</th><th class='p-3 text-left border'>Description</th></tr>
                    <tr><td class='p-3 border font-mono'>username</td><td class='p-3 border'>string</td><td class='p-3 border'>The user's unique login ID.</td></tr>
                    <tr><td class='p-3 border font-mono'>password</td><td class='p-3 border'>string</td><td class='p-3 border'>The secure password provided during registration.</td></tr>
                </table>

                <h4 class='text-slate-900 font-bold mb-4'>Technical In-Depth</h4>
                <p class='mb-6'>The <code>access</code> token is short-lived (60 mins) and should be included in the <code>Authorization: Bearer</code> header. The <code>refresh</code> token is used to obtain new access tokens without re-authenticating.</p>

                {self.get_sdk_tabs('auth-jwt-usage',
                    js_code="import { Puxbay } from '@puxbay/sdk';\n\nconst pux = new Puxbay({ apiKey: 'YOUR_KEY' });\n\n// Login and store tokens\nconst { access, refresh } = await pux.auth.login('johndoe', 'secret');\n\n// Local storage or cookie persistence\nlocalStorage.setItem('p_access', access);\nlocalStorage.setItem('p_refresh', refresh);",
                    py_code="from puxbay import Puxbay\n\nclient = Puxbay(api_key='pb_live_...')\n\n# Perform authentication\ntokens = client.auth.login(username='johndoe', password='secret')\nprint(f'Access Granted: {tokens[\"access\"]}')"
                )}
            """,
            order=1
        )

        DocumentationArticle.objects.create(
            section=sections[('Authentication & Security', 'api')],
            title='Staff PIN Verification',
            slug='auth-pin',
            content=f"""
                <p class='mb-8'>For retail terminal environments, staff can breathe easy with <b>PIN-based authentication</b>. This allows quick switching between cashiers on a shared hardware terminal.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>Implementation Architecture</h4>
                <p class='mb-6'>The client must first authenticate with an <code>X-API-Key</code> representing the branch. Once the branch context is established, a 4-6 digit PIN is sent to obtain a fleeting staff session.</p>

                {self.get_sdk_tabs('auth-pin-usage',
                    js_code="// PIN Login for terminal shared sessions\nconst staffSession = await pux.offline.pinLogin('1234');\n\nconsole.log(`Logged in as: ${staffSession.user.name}`);\nconsole.log(`Role: ${staffSession.user.role}`);",
                    py_code="staff = client.offline.pin_login(pin='1234')\nprint(f'Active Session: {staff[\"user\"][\"username\"]}')"
                )}
            """,
            order=2
        )

        # --- 2. PRODUCT & INVENTORY ---
        DocumentationArticle.objects.create(
            section=sections[('Product & Inventory', 'api')],
            title='Inventory Stock Adjustments',
            slug='inv-adjust',
            content=f"""
                <p class='text-lg mb-8'>Maintain absolute precision in your retail stock levels using the <b>Audit-Logged Stock Adjustment</b> system.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>System Logic</h4>
                <ul class='list-disc pl-6 mb-8 text-slate-600 space-y-2'>
                    <li>Adjustments are relative integers (Positive for restock, negative for shrinkage/loss).</li>
                    <li>Every adjustment triggers a <code>ProductHistory</code> snapshot.</li>
                    <li>Automatic threshold checking for <b>Low Stock Alerts</b>.</li>
                </ul>

                {self.get_sdk_tabs('inv-adjust-usage',
                    js_code="// Adjust stock for a specific SKU\nconst result = await pux.products.adjustStock('prod_uuid', {\n    adjustment: 25,\n    reason: 'Received shipment batch #902'\n});\n\nconsole.log(`New Level: ${result.new_quantity}`);",
                    py_code="result = client.products.adjust_stock(\n    'prod_uuid', \n    adjustment=-2, \n    reason='Damaged during display setup'\n)\nprint(f'Stock updated: {result[\"new_quantity\"]}')"
                )}
            """,
            order=1
        )

        DocumentationArticle.objects.create(
            section=sections[('Product & Inventory', 'api')],
            title='Product Audit History',
            slug='inv-history',
            content=f"""
                <p class='mb-8'>Retrieve an exhaustive timeline of every change made to a product's state, including price shifts, stock movements, and metadata updates.</p>
                
                {self.get_sdk_tabs('inv-history-usage',
                    js_code="// Fetch audit logs\nconst history = await pux.products.history('prod_uuid');\n\nhistory.forEach(event => {\n    console.log(`[${event.changed_at}] ${event.action}: ${event.changes_summary}`);\n});",
                    py_code="history = client.products.history('prod_uuid')\nfor entry in history:\n    print(f\"{entry['changed_at']}: {entry['changes_summary']}\")"
                )}
            """,
            order=2
        )

        # --- 3. CUSTOMER & CRM ---
        DocumentationArticle.objects.create(
            section=sections[('Customer & CRM', 'api')],
            title='Customer CRM Management',
            slug='crm-manage',
            content=f"""
                <p class='text-lg mb-8'>The <b>Customer CRM</b> engine is the central repository for all shopper data, tier assignments, and transaction history.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>Query Filters</h4>
                <p class='mb-6'>You can filter customers by <code>tier</code>, <code>branch</code>, or use the <code>search</code> parameter to look through names, emails, and phones.</p>

                {self.get_sdk_tabs('crm-manage-usage',
                    js_code="// List customers with search\nconst results = await pux.customers.list({ search: 'John', page: 1 });\n\n// Create a new shopper\nconst customer = await pux.customers.create({\n    name: 'Jane Smith',\n    email: 'jane@example.com',\n    phone: '+1234567890'\n});",
                    py_code="customers = client.customers.list(search='John')\nnew_customer = client.customers.create(\n    name='Jane Smith', \n    email='jane@example.com'\n)"
                )}
            """,
            order=1
        )

        DocumentationArticle.objects.create(
            section=sections[('Customer & CRM', 'api')],
            title='Loyalty Engine & Points',
            slug='crm-loyalty',
            content=f"""
                <p class='text-lg mb-8'>Engage your audience with a sophisticated <b>Loyalty Points & Tiers</b> engine.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>Point Accumulation Logic</h4>
                <p class='mb-6'>Points are typically generated at checkout (e.g., 1 point per $1 spent). The API allows for manual overrides for customer service scenarios.</p>

                {self.get_sdk_tabs('crm-loyalty-usage',
                    js_code="// Reward customer points manually\nconst balance = await pux.customers.addLoyaltyPoints('cust_uuid', {\n    points: 500,\n    description: 'Boutique opening bonus'\n});\n\n// Recalculate tier based on lifetime spends\nconst tierUpdate = await pux.customers.refreshTier('cust_uuid');",
                    py_code="client.customers.add_loyalty_points(\n    'cust_uuid', \n    points=100, \n    description='Purchase rebate'\n)\ntier = client.customers.refresh_tier('cust_uuid')\nprint(f'Customer is now in tier: {tier[\"current_tier\"]}')"
                )}
            """,
            order=2
        )

        # --- 4. POINT OF SALE (POS) ---
        DocumentationArticle.objects.create(
            section=sections[('Point of Sale (POS)', 'api')],
            title='The Order Lifecycle',
            slug='pos-order',
            content=f"""
                <p class='text-lg mb-8'>Processing a sale is the heartbeat of the POS. Puxbay supports complex order structures including multi-item checkout, tax calculation, and various payment tenders.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>Standard Transaction Payload</h4>
                <pre class='mb-8'>
{{
    "customer": "uuid",
    "items": [
        {{ "product": "uuid", "quantity": 1, "price": 29.99 }}
    ],
    "payment_method": "cash",
    "status": "completed"
}}
                </pre>

                {self.get_sdk_tabs('pos-order-usage',
                    js_code="// Create a production order\nconst order = await pux.orders.create({\n    customer: 'cust_uuid',\n    items: [\n        { product: 'prod_1', quantity: 2, price: 15.00 },\n        { product: 'prod_2', quantity: 1, price: 99.99 }\n    ],\n    payment_method: 'card',\n    status: 'completed'\n});\n\nconsole.log(`Receipt generated: ${order.order_number}`);",
                    py_code="order = client.orders.create(\n    customer='cust_uuid',\n    items=[\n        {'product': 'prod_id', 'quantity': 1, 'price': 50.0}\n    ],\n    payment_method='transfer'\n)\nprint(f\"Order Successful: {order['id']}\")"
                )}
            """,
            order=1
        )

        # --- 5. SUPPLY CHAIN & LOGISTICS ---
        DocumentationArticle.objects.create(
            section=sections[('Supply Chain & Logistics', 'api')],
            title='Purchase Orders (PO)',
            slug='log-po',
            content=f"""
                <p class='text-lg mb-8'>The <b>Purchase Order</b> system handles incoming shipments from suppliers, automating the restock workflow and financial reconciliation.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>Transition Logic</h4>
                <p class='mb-6'>A PO moves from <code>draft</code> to <code>sent</code>, and finally <code>received</code>. Receiving a PO automatically increments the stock quantities of listed products.</p>

                {self.get_sdk_tabs('log-po-usage',
                    js_code="// Create a PO for a supplier\nconst po = await pux.purchaseOrders.create({\n    supplier: 'supp_uuid',\n    branch: 'branch_uuid',\n    expected_date: '2025-01-20',\n    items: [\n        { product: 'prod_uuid', quantity: 100, cost: 5.50 }\n    ]\n});",
                    py_code="po = client.purchase_orders.create(\n    supplier_id='supp_uuid', \n    items=[{'product': 'p1', 'quantity': 50}]\n)"
                )}
            """,
            order=1
        )

        DocumentationArticle.objects.create(
            section=sections[('Supply Chain & Logistics', 'api')],
            title='Inter-Branch Transfers',
            slug='log-transfers',
            content=f"""
                <p class='text-lg mb-8'>Securely move stock between different retail locations while maintaining real-time visibility and transit tracking.</p>

                {self.get_sdk_tabs('log-transfer-usage',
                    js_code="// Request a stock transfer\nawait pux.stockTransfers.create({\n    source_branch: 'b1_uuid',\n    destination_branch: 'b2_uuid',\n    items: [{ product: 'p1', quantity: 10 }]\n});",
                    py_code="client.stock_transfers.create(\n    source='b1', \n    destination='b2', \n    items=[{'product': 'p1', 'quantity': 5}]\n)"
                )}
            """,
            order=2
        )

        # --- 6. ANALYTICS & INTELLIGENCE ---
        DocumentationArticle.objects.create(
            section=sections[('Analytics & Intelligence', 'api')],
            title='AI Restock Intelligence',
            slug='ai-restock',
            content=f"""
                <p class='text-lg mb-8'>Never miss a sale due to a stockout. Our <b>Predictive Restock Engine</b> analyzes 30-day sales velocity to recommend inventory replenishment levels.</p>
                
                {self.get_sdk_tabs('ai-restock-usage',
                    js_code="// Fetch intelligent restock suggestions\nconst recommendations = await pux.reports.restockRecommendations();\n\nrecommendations.forEach(rec => {\n    console.log(`Product: ${rec.name}`);\n    console.log(`Sales Velocity: ${rec.avg_daily_sales}/day`);\n    console.log(`Recommended Restock: ${rec.recommended_restock} units`);\n});",
                    py_code="recs = client.reports.restock_recommendations()\nfor r in recs:\n    if r['recommended_restock'] > 0:\n        print(f\"Action Required: Order {r['recommended_restock']} of {r['name']}\")"
                )}
            """,
            order=1
        )

        # --- 7. INFRASTRUCTURE & SYNC ---
        DocumentationArticle.objects.create(
            section=sections[('Infrastructure & Sync', 'api')],
            title='Offline Idempotency Sync',
            slug='infra-offline',
            content=f"""
                <p class='text-lg mb-8'>The <b>Offline Sync Protocol</b> ensures data integrity even when the internet is unreliable. Uses idempotent UUID tracking to prevent duplicate transactions.</p>
                
                <h4 class='text-slate-900 font-bold mb-4'>The Sync Workflow</h4>
                <ol class='list-decimal pl-6 mb-8 text-slate-600 space-y-2'>
                    <li>Transaction created on hardware (Offline).</li>
                    <li>UUID generated and stored locally in indexDB/SQLite.</li>
                    <li>Once online, transaction sent to <code>/api/v1/offline/transaction/</code>.</li>
                    <li>Cloud attempts processing based on <code>type</code> (order, transfer, shift_close).</li>
                </ol>

                {self.get_sdk_tabs('infra-sync-usage',
                    js_code="// Perform an idempotent offline sync\nawait pux.offline.sync({ \n    uuid: 'CLIENT_SIDE_UUID_V4',\n    type: 'order',\n    data: {\n        customer: 'cust_id',\n        items: [...]\n    }\n});",
                    py_code="client.offline.sync(\n    uuid='uuid-v4-here',\n    type='complete_order',\n    data={'order_id': 'cloud_id'}\n)"
                )}
            """,
            order=1
        )

        self.stdout.write(self.style.SUCCESS(f'Successfully seeded {DocumentationArticle.objects.count()} IN-DEPTH API endpoints!'))
