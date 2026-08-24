from django.urls import path, include
from .views import documentation_portal

urlpatterns = [
    # Swagger & ReDoc are still here for now if needed, or moved
    path('', documentation_portal, {'doc_type': 'api'}, name='developer_portal'),
    path('reference/', include('documentation.urls')), # This includes Swagger/ReDoc
]
