for i in {0..12}
do
  pulumi stack select zone-"$i"
  shift_amount=$((i+1))
  if (( shift_amount % 13 == 0 ))
  then
      shift_amount=0
  fi
  pulumi config set testerZone $shift_amount
  pulumi up --yes
done