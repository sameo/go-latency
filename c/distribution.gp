#!/usr/bin/gnuplot --persist

clear
reset
set key off
set border 3


set style fill solid 1.0
set title "C scheduling latency (2000 runs)"
set ylabel "Occurences"
set xlabel "Latency (µs)"

bin_width = 0.1;
bin_number(x) = floor(x/bin_width)
rounded(x) = bin_width * ( bin_number(x) + 0.1 )

file=ARG1

plot file using (rounded($1)):(1) smooth frequency with impulses
pause -1