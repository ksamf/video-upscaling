FROM node:22.13.0

COPY package.json package-lock.json ./
COPY . .
RUN  npm install

RUN npm run build

EXPOSE 4173

CMD ["npm", "run", "preview"]