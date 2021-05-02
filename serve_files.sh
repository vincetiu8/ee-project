mkdir temp
cd temp || exit
head -c 1M </dev/urandom > 1M
head -c 10M </dev/urandom > 10M
head -c 100M </dev/urandom > 100M
head -c 1G </dev/urandom > 1G

sudo apt -y install imagemagick
sudo apt -y install libav-tools
mx=1000;my=1000;head -c "$((3*mx*my))" /dev/urandom | convert -depth 8 -size "${mx}x${my}" RGB:- 1000px.png
mx=10000;my=10000;head -c "$((3*mx*my))" /dev/urandom | convert -depth 8 -size "${mx}x${my}" RGB:- 10000px.png
mx=100; my=100; nframes=1000; dd if=/dev/urandom bs="$((mx*my*3))" count="$nframes" | avconv -r 25 -s "${mx}x${my}" -f rawvideo -pix_fmt rgb24 -i - 100px.mp4
mx=1000; my=1000; nframes=1000; dd if=/dev/urandom bs="$((mx*my*3))" count="$nframes" | avconv -r 25 -s "${mx}x${my}" -f rawvideo -pix_fmt rgb24 -i - 1000px.mp4
mkdir 10F
for _ in {0..9}
do
  head -c 100M </dev/urandom > 10F/100M
done
echo "10F 1M 10M 100M 1G 1000px.png 10000px.png 100px.mp4 1000px.mp4" > files.txt
until ipfs add 1M
do
  sleep 1
done
ipfs add 10F -r | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 1M  | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 10M | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 100M | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 1G | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 1000px.png | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 10000px.png | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 100px.mp4 | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add 1000px.mp4 | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
ipfs add files.txt | sed "s/[^ ]* //" | sed "s/.[^ ]*$//" | tee -a files.txt
sudo chown -R ubuntu .
nohup python3 -m http.server 8000 &