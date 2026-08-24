from django.urls import path
from .views import documentation_portal

urlpatterns = [
    path('', documentation_portal, {'doc_type': 'manual'}, name='manual_portal'),
]
