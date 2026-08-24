import os
import django
from django.conf import settings
from django.test import RequestFactory
from possystem.dashboard import admin_dashboard_view
from django.contrib.auth.models import User

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'possystem.settings')
django.setup()

factory = RequestFactory()
request = factory.get('/puxbay-hq/dashboard/')
user = User.objects.filter(is_staff=True).first()
if not user:
    user = User.objects.create_superuser('temp_admin', 'admin@example.com', 'password')
request.user = user

try:
    response = admin_dashboard_view(request)
    print(f"Status Code: {response.status_code}")
    print(f"Content Length: {len(response.content)}")
    if response.status_code != 200:
        print(f"Content: {response.content.decode()[:500]}")
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
