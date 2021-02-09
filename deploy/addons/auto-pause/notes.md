## Running in docker 
```
docker run -p 38443:38443 -v $(pwd):/t -it --rm --name haproxy haproxy -f /t/haproxy.cfg 
```