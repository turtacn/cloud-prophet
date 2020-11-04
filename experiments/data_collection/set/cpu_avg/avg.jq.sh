printf  "$1 " >  $1.value    
jq  '[.[].dps[]] | add/length' $1 >> $1.value
