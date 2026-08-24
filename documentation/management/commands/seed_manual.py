from django.core.management.base import BaseCommand
from django.utils.text import slugify
from documentation.models import DocumentationSection, DocumentationArticle

class Command(BaseCommand):
    help = 'Seeds the database with the Ultimate Exhaustive User Manual'

    def handle(self, *args, **options):
        self.stdout.write('Deploying Ultimate Exhaustive User Manual...')

        # Wipe existing manual records to ensure a clean, structured rebuild
        DocumentationSection.objects.filter(doc_type='manual').delete()

        # The Master Data Structure
        data = [
            {
                'title': '1. System Architecture & Setup',
                'icon': 'architecture',
                'order': 1,
                'articles': [
                    {
                        'title': 'The Puxbay Hierarchy',
                        'content': """
                            <h1>Understanding Your Enterprise</h1>
                            <p>Puxbay is structured into three distinct layers to ensure data isolation and operational efficiency.</p>
                            <ul>
                                <li><strong>Tenant:</strong> Your top-level organizational unit. Each tenant has a unique subdomain (e.g., <code>tech-mart.puxbay.com</code>).</li>
                                <li><strong>Branches:</strong> Individual physical or virtual store locations. All sales, inventory, and staff are attached to a specific branch.</li>
                                <li><strong>Users:</strong> Your employees. Each user is assigned to one or more branches with a specific role.</li>
                            </ul>
                        """
                    },
                    {
                        'title': 'Mastering Subdomains',
                        'content': """
                            <h1>DNS & Domain Management</h1>
                            <p>Puxbay uses a multi-tenant DNS architecture:</p>
                            <ul>
                                <li><strong>Landing Page:</strong> <code>www.puxbay.com</code> (Global marketing and account creation).</li>
                                <li><strong>Business Dashboard:</strong> <code>[your-subdomain].puxbay.com</code> (Your private operations center).</li>
                                <li><strong>API Gateway:</strong> <code>api.puxbay.com</code> (Used by external apps, mobile terminals, and kiosks).</li>
                            </ul>
                        """
                    },
                ]
            },
            {
                'title': '2. Inventory & Procurement',
                'icon': 'inventory_2',
                'order': 2,
                'articles': [
                    {
                        'title': 'Advanced Cataloging',
                        'content': """
                            <h1>Managing Complex Inventory</h1>
                            <p>Puxbay supports advanced inventory types beyond simple products:</p>
                            <ul>
                                <li><strong>Variants:</strong> Manage color, size, or material variations under one SKU.</li>
                                <li><strong>Composites (Bundles):</strong> Group existing products into sets (e.g., "Gift Box"). Selling a bundle automatically reduces stock of each component.</li>
                                <li><strong>Global Products:</strong> Products marked as "Global" appear in all branches, simplifying setup for multi-store franchises.</li>
                            </ul>
                        """
                    },
                    {
                        'title': 'Supply Chain & Procurement',
                        'content': """
                            <h1>Supplier Relations</h1>
                            <p>Move beyond manual restocking with the Procurement module.</p>
                            <ul>
                                <li><strong>Supplier Management:</strong> Track tax IDs, payment terms, and credit limits for your vendors.</li>
                                <li><strong>Purchase Orders (PO):</strong> Create digitised requests. Upon "Receiving" a PO, your stock is updated instantly across the system.</li>
                                <li><strong>Supplier Debt:</strong> If you buy on credit, the system tracks exactly how much is owed to each vendor in the <strong>Accounts Payable</strong> ledger.</li>
                            </ul>
                        """
                    },
                ]
            },
            {
                'title': '3. Point of Sale (POS)',
                'icon': 'point_of_sale',
                'order': 3,
                'articles': [
                    {
                        'title': 'The Professional POS',
                        'content': """
                            <h1>High-Velocity Sales</h1>
                            <p>The Puxbay POS is an industry-leading selling terminal optimized for any device.</p>
                            <ul>
                                <li><strong>Offline Persistence:</strong> If the internet fails, sales are cached in the browser and synced immediately when connectivity is restored.</li>
                                <li><strong>Multi-Channel Payments:</strong> Accept Cash, Card, Mobile Wallets, or **Store Credit**.</li>
                                <li><strong>Quick Switch:</strong> Cashiers can switch shifts using a secure 4-6 digit **POS PIN**, preventing unauthorized access while maintaining speed.</li>
                            </ul>
                        """
                    },
                    {
                        'title': 'Service & Returns',
                        'content': """
                            <h1>After-Sales Workflow</h1>
                            <p>Process returns directly from the transaction log. You can refund to the original payment method, issue **Store Credit**, or provide Cash. The system handles inventory restocking and tax reversals automatically.</p>
                        """
                    },
                ]
            },
            {
                'title': '4. Self-Service Kiosks',
                'icon': 'potted_plant',
                'order': 4,
                'articles': [
                    {
                        'title': 'Setting Up Kiosks',
                        'content': """
                            <h1>Modern Customer Experience</h1>
                            <p>Puxbay allows you to deploy self-service kiosks at your physical branches.</p>
                            <ul>
                                <li><strong>URL-Based Deployment:</strong> Kiosks run on a specialized Vue.js interface accessible at <code>[tenant].puxbay.com/kiosk/[branch-id]/</code>.</li>
                                <li><strong>Security:</strong> Kiosks are locked to the branch ID and cannot access administrative dashboards.</li>
                            </ul>
                        """
                    },
                    {
                        'title': 'Kiosk Workflow',
                        'content': """
                            <h1>Self-Checkout Journey</h1>
                            <p>1. <strong>Browse:</strong> Customers select items from the visual catalog.</p>
                            <p>2. <strong>Pay:</strong> Customers pay via integrated card readers or mobile money gateways.</p>
                            <p>3. <strong>Ready:</strong> The POS creates a 'Kiosk Order'. Once staff prepares the order, the status is updated to 'Ready' for customer pickup.</p>
                        """
                    },
                ]
            },
            {
                'title': '5. E-commerce & Storefront',
                'icon': 'shopping_cart',
                'order': 5,
                'articles': [
                    {
                        'title': 'Online Presence',
                        'content': """
                            <h1>Global Storefront</h1>
                            <p>Every Puxbay tenant gets a built-in E-commerce storefront to reach customers anywhere.</p>
                            <ul>
                                <li><strong>Custom Branding:</strong> Upload banners, logos, and select your brand's primary color in **Storefront Settings**.</li>
                                <li><strong>Flexible Fulfillment:</strong> Allow In-Store Pickup or Local Delivery with automated fee calculations.</li>
                            </ul>
                        """
                    },
                    {
                        'title': 'Customer Engagement',
                        'content': """
                            <h1>Driving Conversion</h1>
                            <ul>
                                <li><strong>Wishlists:</strong> Allow customers to save items for later.</li>
                                <li><strong>Product Reviews:</strong> Build trust with star ratings and verified customer comments.</li>
                                <li><strong>Abandoned Cart Recovery:</strong> The system identifies shoppers who left items in their cart and can send automated nudge emails to recover the sale.</li>
                                <li><strong>Coupons:</strong> Manage discount codes with specific expiration dates and minimum purchase requirements.</li>
                            </ul>
                        """
                    },
                ]
            },
            {
                'title': '6. Staff & Attendance',
                'icon': 'badge',
                'order': 6,
                'articles': [
                    {
                        'title': 'Workforce Management',
                        'content': """
                            <h1>Roles & Accountability</h1>
                            <p>Puxbay supports granular roles (Admin, Manager, Procurement, Sales, Financial). Each role has strictly enforced permissions (e.g., Sales can't delete products).</p>
                        """
                    },
                    {
                        'title': 'Attendance System',
                        'content': """
                            <h1>Tracking Presence</h1>
                            <p>The **Attendance Dashboard** allows employees to clock-in/out. Admins receive real-time visibility into who is currently working and can generate duration reports for monthly payroll calculations.</p>
                        """
                    },
                ]
            },
            {
                'title': '7. CRM & Loyalty',
                'icon': 'rewarded_ads',
                'order': 7,
                'articles': [
                    {
                        'title': 'VIP & Loyalty Tiers',
                        'content': """
                            <h1>Reward Your Best Customers</h1>
                            <p>Automatically categorize customers into Tiers (Silver, Gold, Platinum) based on their total spend. Each tier can have a pre-configured discount that applies automatically at the check-out.</p>
                        """
                    },
                    {
                        'title': 'Smart Marketing',
                        'content': """
                            <h1>Automated Campaigns</h1>
                            <p>Reach your customers via SMS or Email using **Triggers**:</p>
                            <ul>
                                <li><strong>Inactivity:</strong> Nudge customers who haven't bought in 30 days.</li>
                                <li><strong>Birthdays:</strong> Send automated personalized discount codes.</li>
                                <li><strong>First-Purchase:</strong> Welcome new users after their first successful transaction.</li>
                            </ul>
                        """
                    },
                ]
            },
            {
                'title': '8. Financials & Taxes',
                'icon': 'account_balance',
                'order': 8,
                'articles': [
                    {
                        'title': 'Profit & Loss Logic',
                        'content': """
                            <h1>Real-time Accounting</h1>
                            <p>The P&L statement is generated in real-time by subtracting **COGS** (at the time of sale) and **Operating Expenses** (Rent, Salaries) from your **Gross Revenue**.</p>
                        """
                    },
                    {
                        'title': 'Tax Management',
                        'content': """
                            <h1>Regulatory Compliance</h1>
                            <p>Configure VAT, Sales Tax, or GST in the **Tax Configuration** panel. This ensures every POS and E-commerce transaction includes the correct tax burden, and generates accurate exportable reports for filing.</p>
                        """
                    },
                ]
            },
            {
                'title': '9. SaaS Billing & Subscriptions',
                'icon': 'credit_card',
                'order': 9,
                'articles': [
                    {
                        'title': 'Managing Your Subscription',
                        'content': """
                            <h1>Plans and Quotas</h1>
                            <p>Puxbay is a subscription service. Your plan determines your limits for **Branches**, **User Seats**, and **API Requests per Day**.</p>
                            <p>Invoices and payment methods (Card/Regional) are managed directly within the **Billing Cabinet**.</p>
                        """
                    },
                ]
            },
            {
                'title': '10. Security & Notifications',
                'icon': 'admin_panel_settings',
                'order': 10,
                'articles': [
                    {
                        'title': 'Alerting & Monitoring',
                        'content': """
                            <h1>Stay Informed</h1>
                            <p>The **Notification Center** tracks critical events across categories:</p>
                            <ul>
                                <li><strong>Inventory:</strong> Low stock or out-of-stock alerts.</li>
                                <li><strong>Security:</strong> Notifications for large refunds, voids, or unusual login activity.</li>
                                <li><strong>System:</strong> Backup statuses and update completion alerts.</li>
                            </ul>
                        """
                    },
                    {
                        'title': 'Audit Logs',
                        'content': """
                            <h1>The Ultimate Trial</h1>
                            <p>Admins can access the **Global Audit Log**, which records every single data change with the associated User, Timestamp, and IP Address. This is your primary defense against internal fraud.</p>
                        """
                    },
                ]
            },
        ]

        for sec_data in data:
            section = DocumentationSection.objects.create(
                title=sec_data['title'],
                icon=sec_data['icon'],
                order=sec_data['order'],
                doc_type='manual'
            )
            self.stdout.write(f"Created Section: {section.title}")

            for i, art_data in enumerate(sec_data['articles']):
                DocumentationArticle.objects.create(
                    section=section,
                    title=art_data['title'],
                    slug=slugify(f"manual-{section.title}-{art_data['title']}"),
                    content=art_data['content'],
                    order=i + 1,
                    is_published=True
                )
                self.stdout.write(f"  - Created Article: {art_data['title']}")

        self.stdout.write(self.style.SUCCESS('Universal System-Wide User Manual (10 Master Modules) seeded successfully!'))
