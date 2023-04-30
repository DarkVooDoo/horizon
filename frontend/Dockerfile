FROM node:19-alpine3.17
WORKDIR /app
RUN mkdir frontend
WORKDIR /app/frontend
COPY package.json .
RUN npm install
COPY . .
RUN npm run build
ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000
CMD node ./dist/server/entry.mjs