from django.shortcuts import render
from .models import DocumentationSection, DocumentationArticle

def documentation_portal(request, doc_type='api'):
    """
    Main portal view for documentation.
    Displays sections and articles filtered by doc_type.
    """
    sections = DocumentationSection.objects.prefetch_related('articles').filter(
        doc_type=doc_type,
        articles__is_published=True
    ).distinct()
    
    query = request.GET.get('q')
    if query:
        # Simple search across articles
        from django.db.models import Q
        search_results = DocumentationArticle.objects.filter(
            Q(title__icontains=query) | Q(content__icontains=query),
            section__doc_type=doc_type,
            is_published=True
        ).select_related('section')
        
        return render(request, 'documentation/portal.html', {
            'search_results': search_results,
            'query': query,
            'title': f'Search Results: {query}'
        })

    context = {
        'sections': sections,
        'title': 'API Reference' if doc_type == 'api' else 'User Manual'
    }
    return render(request, 'documentation/portal.html', context)
