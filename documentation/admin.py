from django.contrib import admin

from .models import DocumentationSection, DocumentationArticle

class DocumentationArticleInline(admin.TabularInline):
    model = DocumentationArticle
    extra = 1
    fields = ('title', 'slug', 'order', 'is_published')
    prepopulated_fields = {'slug': ('title',)}

@admin.register(DocumentationSection)
class DocumentationSectionAdmin(admin.ModelAdmin):
    list_display = ('title', 'order', 'article_count')
    list_editable = ('order',)
    search_fields = ('title',)
    inlines = [DocumentationArticleInline]
    
    def article_count(self, obj):
        return obj.articles.count()
    article_count.short_description = 'Articles'

from tinymce.widgets import TinyMCE
from django.db import models

@admin.register(DocumentationArticle)
class DocumentationArticleAdmin(admin.ModelAdmin):
    list_display = ('title', 'section', 'order', 'is_published', 'updated_at')
    list_filter = ('section', 'is_published', 'updated_at')
    search_fields = ('title', 'content')
    list_editable = ('order', 'is_published')
    prepopulated_fields = {'slug': ('title',)}
    autocomplete_fields = ['section']

    class Media:
        js = ('/static/tinymce/tinymce.min.js',)
