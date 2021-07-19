if [ "$#" -ne 1 ]; then
          exit 1
fi
until curl "http://$1:8000/files.txt"
do
  sleep 1
done
files=$(curl "http://$1:8000/files.txt" -s)
readarray -t file_hashes <<< "$files"
until ipfs pin add "${file_hashes[11]}"
do
  sleep 1
done
for file in "${file_hashes[@]:12}"
do
  ipfs pin add "$file"
done
