# Stage 1: Build the React application
FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .


RUN npm run build

# Stage 2: Serve the application with Nginx
FROM nginx:alpine
RUN rm -rf /usr/share/nginx/html/*
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY generate-env.sh /docker-entrypoint.d/99-generate-env.sh
RUN chmod +x /docker-entrypoint.d/99-generate-env.sh

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
