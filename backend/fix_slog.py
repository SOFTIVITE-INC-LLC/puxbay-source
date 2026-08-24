import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if 'zap.' in content:
        content = content.replace('zap.String(', 'slog.String(')
        content = content.replace('zap.Int(', 'slog.Int(')
        content = content.replace('zap.Any(', 'slog.Any(')
        content = content.replace('zap.Duration(', 'slog.Any("duration", ')
        content = content.replace('zap.Error(', 'slog.Any("error", ')
        
        # also remove "go.uber.org/zap"
        content = re.sub(r'"go.uber.org/zap"', '', content)
        if 'log/slog' not in content:
            content = content.replace('import (', 'import (\n\t"log/slog"')
            
        with open(filepath, 'w') as f:
            f.write(content)

for root, dirs, files in os.walk('.'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
