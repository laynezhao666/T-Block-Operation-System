# 第一阶段：构建应用
FROM node:14 as builder

WORKDIR /app
COPY . /app

RUN npm install --registry=https://registry.npmjs.org/
RUN npm run public-tedge
RUN cp -r dist/main/adaptor.html dist/main/index.html

# 第二阶段：构建最终镜像
FROM nginx:latest

COPY ./nginx_docker.conf /etc/nginx/nginx.conf.template
COPY --from=builder /app/dist /usr/share/nginx/html/tnebula

RUN rm /etc/nginx/conf.d/default.conf

WORKDIR /usr/share/nginx/html

EXPOSE 80

# 启动时用envsubst替换环境变量，动态提取所有PORT_开头的变量，避免影响nginx自身的$变量
CMD ["/bin/sh", "-c", "envsubst \"$(printf '${%s} ' $(env | grep '^PORT_' | cut -d= -f1))\" < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf && nginx -g 'daemon off;'"]
