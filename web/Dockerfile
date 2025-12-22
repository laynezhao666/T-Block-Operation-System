# 第一阶段：构建应用
FROM node:14 as builder

WORKDIR /app
COPY . /app

RUN npm install --registry=https://registry.npmjs.org/
RUN npm run public-tedge
RUN cp -r dist/main/adaptor.html dist/main/index.html

# 第二阶段：构建最终镜像
FROM nginx:latest

COPY ./nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html/tnebula

RUN rm /etc/nginx/conf.d/default.conf

WORKDIR /usr/share/nginx/html

EXPOSE 80

CMD ["nginx","-g","daemon off;"]
