# フロントエンド（Next.js）の Dockerfile
#   dev   ：ローカル開発用。next dev を 3000 番で起動する（compose.yaml から使う）
#   build ：静的エクスポートを作る。/app/out にビルド結果ができる
#           #31 で Go のイメージを作るときに、このステージから out/ をコピーする

FROM node:24-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM deps AS dev
ENV NEXT_TELEMETRY_DISABLED=1
COPY . .
EXPOSE 3000
CMD ["npm", "run", "dev", "--", "--hostname", "0.0.0.0"]

FROM deps AS build
ENV NEXT_TELEMETRY_DISABLED=1
COPY . .
RUN npm run build

# docker build --target out --output type=local,dest=out . で、ホストに out/ を書き出せる
FROM scratch AS out
COPY --from=build /app/out /
