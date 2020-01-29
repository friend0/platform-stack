i=0;
while :; do
    if [ $i -lt 5 ]
    then
         echo "[$(uname -n)] $(date)";
         i=$((i+1));
    fi
    sleep 1;
done;