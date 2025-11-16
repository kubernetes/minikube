FROM node:slim
RUN apt-get update && apt-get install curl -y
EXPOSE 8080
COPY server.js .
CMD [ "node", "server.js" ]
