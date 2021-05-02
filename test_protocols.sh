until curl "http://$1:8000/files.txt"
do
  sleep 1
done
files=$(curl "http://$1:8000/files.txt" -s)
readarray -t file_data <<< "$files"
file_names=(${file_data[0]})
file_hashes=("${file_data[@]:1}")
echo "File names:" "${file_names[@]}"
echo "File hashes:" "${file_hashes[@]}"
num_files=${#file_names[@]}
echo "Num files:" ${#file_names[@]}
ipfs get "${file_hashes[-1]}"
rm "${file_hashes[-1]}"
file_name="$3-x$2.txt"
echo "$file_name" > "$file_name"
for ((index=0; index < $num_files; index++))
do
  (time -p bash -c "for _ in {1..$2}; do wget -r -np \"http://$1:8000/${file_names[$index]}\" &> /dev/null; done") 2>&1 | grep -oE "[^[:space:]]+$" | tr "\n" "\t" >> "$file_name"
done
for ((index=0; index < $num_files; index++))
do
  (time -p bash -c "for _ in {1..$2}; do ipfs get ${file_hashes[$index]} &> /dev/null; done") 2>&1 | grep -oE "[^[:space:]]+$" | tr "\n" "\t" >> "$file_name"
  rm -rf "${file_hashes[$index]}"
done

cat "$file_name"

aws s3 cp "$file_name" s3://ipfs-output-bucket