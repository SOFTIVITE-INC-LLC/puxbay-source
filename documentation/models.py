from django.db import models
import uuid

class DocumentationSection(models.Model):
    """Sections for grouping documentation articles (e.g., 'Getting Started', 'Inventory')"""
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    title = models.CharField(max_length=100)
    doc_type = models.CharField(
        max_length=20,
        choices=[('manual', 'User Manual'), ('api', 'API Reference')],
        default='api',
        help_text="Type of documentation this section belongs to"
    )
    order = models.PositiveIntegerField(default=0, help_text="Order in sidebar")
    icon = models.CharField(max_length=50, blank=True, help_text="Material Symbol icon name")
    
    class Meta:
        ordering = ['order', 'title']
    
    def __str__(self):
        return self.title

from tinymce.models import HTMLField

class DocumentationArticle(models.Model):
    """Individual documentation articles"""
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    section = models.ForeignKey(DocumentationSection, on_delete=models.CASCADE, related_name='articles')
    title = models.CharField(max_length=200)
    slug = models.SlugField(max_length=200, unique=True)
    content = HTMLField(help_text="Rich HTML content for the article")
    order = models.PositiveIntegerField(default=0)
    is_published = models.BooleanField(default=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    class Meta:
        ordering = ['section', 'order', 'title']
    
    def __str__(self):
        return f"{self.title} ({self.section.title})"
