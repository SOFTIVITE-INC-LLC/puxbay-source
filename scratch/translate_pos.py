import re

with open('/home/afari/Projects/development/puxbay/branches/templates/branches/pos.html', 'r') as f:
    content = f.read()

# Remove Django template tags
content = re.sub(r'\{% extends .*? %\}', '', content)
content = re.sub(r'\{% load .*? %\}', '', content)
content = re.sub(r'\{% block .*? %\}', '', content)
content = re.sub(r'\{% endblock %\}', '', content)
content = re.sub(r'\{% verbatim %\}', '', content)
content = re.sub(r'\{% endverbatim %\}', '', content)
content = re.sub(r'\{% url .*? %\}', '#', content)
content = re.sub(r'\{% static \'([^\']+)\' %\}', r'assets/\1', content)

# Remove script blocks and styles at the end
content = re.sub(r'<script.*?</script>', '', content, flags=re.DOTALL)
# Keep styles inside the head or just keep them inline

# Replace Vue directives
content = re.sub(r'v-if="([^"]+)"', r'@if (\1)', content)
# Wait, @if in Angular 17 is block syntax, it wraps the element. 
# Using *ngIf is safer for inline replacement. Let's use *ngIf and *ngFor
content = re.sub(r'v-if="([^"]+)"', r'*ngIf="\1"', content)
content = re.sub(r'v-show="([^"]+)"', r'[class.hidden]="!(\1)"', content)

# v-for="item in items" :key="item.id" -> *ngFor="let item of items; trackBy: trackByFn"
# Simple regex for v-for="item in items"
content = re.sub(r'v-for="\(([^,]+),\s*([^\)]+)\)\s+in\s+([^"]+)"', r'*ngFor="let \1 of \3; let \2 = index"', content)
content = re.sub(r'v-for="([^ ]+)\s+in\s+([^"]+)"\s+:key="[^"]+"', r'*ngFor="let \1 of \2"', content)
content = re.sub(r'v-for="([^ ]+)\s+in\s+([^"]+)"', r'*ngFor="let \1 of \2"', content)

# Event bindings @click -> (click)
content = re.sub(r'@click="([^"]+)"', r'(click)="\1"', content)

# v-model -> [(ngModel)]
content = re.sub(r'v-model(?:.number)?="([^"]+)"', r'[(ngModel)]="\1"', content)

# :class="['a', condition ? 'b' : 'c']" -> [ngClass]="['a', condition ? 'b' : 'c']"
content = re.sub(r':class="([^"]+)"', r'[ngClass]="\1"', content)

# :src -> [src]
content = re.sub(r':src="([^"]+)"', r'[src]="\1"', content)

# :disabled -> [disabled]
content = re.sub(r':disabled="([^"]+)"', r'[disabled]="\1"', content)

# State mapping
content = content.replace('state.', 'facade.')
content = content.replace('facade.currency', 'facade.currency()')
content = content.replace('facade.searchQuery', 'facade.searchQuery()')
# We have to be careful with () for signals, but let's just do a basic replace and fix manually if needed.
# Since it's too complex to add () to all signals via regex, I'll let Angular compile and I'll fix errors.

# Write to the destination
with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'w') as f:
    f.write(content.strip())
