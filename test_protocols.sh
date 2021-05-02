until curl "http://$1:8000/files.txt"
do
  sleep 1
done
files=$(curl "http://$1:8000/files.txt" -s)
readarray -t file_hashes <<< "$files"
ipfs get "${file_hashes[4]}"
rm "${file_hashes[4]}"
file_name="$3-x$2.txt"
echo "$file_name" > "$file_name"
file_sizes=(1M 10M 100M 1G)
for index in {0..3}
do
  (time -p bash -c "for _ in {1..$2}; do curl \"http://$1:8000/${file_sizes[index]}\" -o /dev/null -s; done") 2>&1 | grep -oE "[^[:space:]]+$" | tr "\n" "\t" >> "$file_name"
done
for index in {0..3}
do
  (time -p bash -c "for _ in {1..$2}; do ipfs get ${file_hashes[$index]} &> /dev/null; done") 2>&1 | grep -oE "[^[:space:]]+$" | tr "\n" "\t" >> "$file_name"
  rm "${file_hashes[$index]}"
done

cat "$file_name"

aws s3 cp "$file_name" s3://ipfs-output-bucket