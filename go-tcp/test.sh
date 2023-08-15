sleep 600
./dockerrm.sh
docker image rm tcp:latest
docker build -t tcp:latest .

SLOWDOWNS=(0 1 2)
rm -rf data
for i in "${SLOWDOWNS[@]}"
do
    mkdir -p data/slowdowns$i
    dir="data/slowdowns$i"
    ./docker.sh
    sleep 10
    ./slowdowns$i.sh &
    # cd /Users/noahcui/Desktop/githubrepos.nosync/YCSB

    # ./bin/ycsb run my -s -P workloads/goleveldb/workloada/2_6 -threads 64 -target 2000 >/Users/noahcui/Desktop/githubrepos.nosync/sosp23-paper28/go-tcp/$dir/ycsb 2>//Users/noahcui/Desktop/githubrepos.nosync/sosp23-paper28/go-tcp/$dir/ycsb_10sec

    # cd /Users/noahcui/Desktop/githubrepos.nosync/sosp23-paper28/go-tcp
    ./bin/benchmarker -c benchmarker/dockerconfig.json -dir data/aug13_$i 

    docker cp s0:/go/src/gotcp/bufferinfo_0_0.csv ./$dir/
    docker cp s1:/go/src/gotcp/bufferinfo_0_1.csv ./$dir/
    docker cp s2:/go/src/gotcp/bufferinfo_0_2.csv ./$dir/

    ./dockerrm.sh

    ./draw.py -input $dir/bufferinfo_0_1.csv -y forwardings -outfile $dir/s1forwarding.pdf
    ./draw.py -input $dir/bufferinfo_0_2.csv -y forwardings -outfile $dir/s2forwarding.pdf

    ./draw.py -input $dir/bufferinfo_0_1.csv -y forwarding_latencies -outfile $dir/s1forwarding_latencies.pdf
    ./draw.py -input $dir/bufferinfo_0_2.csv -y forwarding_latencies -outfile $dir/s2forwarding_latencies.pdf

    sleep 600
done

scp -r data VM:~/.