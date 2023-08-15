docker run -i -p 10001:10001 -d --cpus=1 --cpuset-cpus=0 --memory=512m --memory-swap=1024m --name s0 tcp:latest bin/server -c config/dockerconfig.json -id 0 
docker run -i -p 11001:10001 -d --cpus=1 --cpuset-cpus=1 --memory=512m --memory-swap=1024m --name s1 tcp:latest bin/server -c config/dockerconfig.json -id 1
docker run -i -p 12001:10001 -d --cpus=1 --cpuset-cpus=2 --memory=512m --memory-swap=1024m --name s2 tcp:latest bin/server -c config/dockerconfig.json -id 2


# docker run -i -p 10001:10001 --cpus=1 --cpuset-cpus=0 --memory=512m --memory-swap=1024m --name s0 tcp:latest bin/server -c config/config.json -id 0 -d
# docker run -i -p 11001:10001 --cpus=1 --cpuset-cpus=0 --memory=512m --memory-swap=1024m --name s1 tcp:latest bin/server -c config/config.json -id 1 -d
# docker run -i -p 12001:10001 --cpus=1 --cpuset-cpus=0 --memory=512m --memory-swap=1024m --name s2 tcp:latest bin/server -c config/config.json -id 2 -d