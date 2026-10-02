FROM node:22-alpine AS build

WORKDIR /app

COPY package.json ./
COPY frontend/package.json ./frontend/

RUN npm install

COPY frontend ./frontend

ARG PUBLIC_API_BASE_URL=/api/v1
ENV PUBLIC_API_BASE_URL=${PUBLIC_API_BASE_URL}
ENV NODE_ENV=production

RUN npm run build -w frontend

FROM node:22-alpine AS production

WORKDIR /app

ARG PUBLIC_API_BASE_URL=/api/v1
ENV PUBLIC_API_BASE_URL=${PUBLIC_API_BASE_URL}
ENV NODE_ENV=production
ENV PORT=5000
ENV HOST=0.0.0.0

COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/frontend/package.json ./package.json
COPY --from=build /app/frontend/build ./build

EXPOSE 5000

CMD ["node", "build"]
