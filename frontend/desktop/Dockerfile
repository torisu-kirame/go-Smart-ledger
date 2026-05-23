# Vue3 构建 + Nginx 托管（/api 反代至 gateway-api）
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json .npmrc ./
RUN npm install --cache .npm-cache

COPY index.html vite.config.js ./
COPY public ./public
COPY src ./src

RUN npm run build

FROM nginx:1.27-alpine

RUN apk add --no-cache wget

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=5s --retries=6 --start-period=10s \
  CMD wget -q -O /dev/null http://127.0.0.1/ || exit 1
