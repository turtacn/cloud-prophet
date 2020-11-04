printf  "$1 " >  $1.value    
jq  '[.[].dps[]] | max ' $1 >> $1.value
