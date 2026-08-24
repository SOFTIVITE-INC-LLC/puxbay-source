import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Replacements
    content = content.replace('logger.Log.Fatal', 'logFatal')
    
    # Sugar().Infof(...) -> Info(fmt.Sprintf(...))
    content = re.sub(r'logger\.Log\.Sugar\(\)\.Infof\((.*?)\)', r'logger.Log.Info(fmt.Sprintf(\1))', content)
    content = re.sub(r'logger\.Log\.Sugar\(\)\.Errorf\((.*?)\)', r'logger.Log.Error(fmt.Sprintf(\1))', content)
    content = re.sub(r'logger\.Log\.Sugar\(\)\.Warnf\((.*?)\)', r'logger.Log.Warn(fmt.Sprintf(\1))', content)
    content = re.sub(r'logger\.Log\.Sugar\(\)\.Fatalf\((.*?)\)', r'logFatal(fmt.Sprintf(\1))', content)
    
    # Sugar().Infow(msg, keysAndValues...) -> Info(msg, slog.Any...)
    # This is tricky with regex, we can just leave it or replace manually.
    # Looking at cors.go:
    # logger.Log.Sugar().Errorw("HTTP Request", "status", ... ) -> logger.Log.Error("HTTP Request", "status", ...)
    content = re.sub(r'logger\.Log\.Sugar\(\)\.Errorw\(', r'logger.Log.Error(', content)
    content = re.sub(r'logger\.Log\.Sugar\(\)\.Infow\(', r'logger.Log.Info(', content)
    
    with open(filepath, 'w') as f:
        f.write(content)

for root, dirs, files in os.walk('.'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
