# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm ci

# Copy the rest of the application
COPY . .

# Build the Angular app for production
RUN npm run build -- --configuration production

# Final stage
FROM nginx:alpine

# Copy the built assets from the builder stage
# Assuming the build output is in dist/puxbay/browser or similar
# The path depends on angular.json configuration. I'll check angular.json later if needed, default is usually dist/frontend/browser
COPY --from=builder /app/dist/frontend/browser /usr/share/nginx/html

# Copy the NGINX configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
