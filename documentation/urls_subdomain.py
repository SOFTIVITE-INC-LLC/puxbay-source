from django.urls import path
from .views import documentation_portal

urlpatterns = [
    path('', documentation_portal, name='doc_portal'),
]
