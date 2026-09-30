FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.9 --activate
COPY package.json ./
RUN pnpm install --frozen-lockfile=false
COPY . .
RUN pnpm build
FROM node:22-alpine
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.9 --activate
COPY package.json ./
RUN pnpm install --prod --frozen-lockfile=false
COPY --from=build /app/dist dist
EXPOSE 3001
CMD ["node","dist/main.js"]
