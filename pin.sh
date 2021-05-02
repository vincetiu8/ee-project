if [ "$#" -ne 2 ]; then
          exit 1
fi
files=$(curl "http://$1:8000/files.txt" -s)
readarray -t file_hashes <<< "$files"
for file in "${file_hashes[@]}"
do
  ipfs pin "$file"
done
