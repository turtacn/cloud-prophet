mkdir -p  data_processing && /usr/bin/rm -rf data_processing/*
cp mem_*/*.data data_processing/
cp cpu_*/*.data data_processing/
find data_processing/*.data   -exec  sed  -i    '/ $/d'  {} \;
find data_processing/*.data   -exec  sed  -i    '/null$/d'  {} \;


