#!/usr/bin/gnuplot --persist

clear
reset
set key off
set border 3

set boxwidth 5 absolute
set style fill solid 1.0
set title "Golang scheduling latency"
set ylabel "Occurences"
set xlabel "Latency (µs)"

bin_width = 5;
bin_number(x) = floor(x/bin_width)
rounded(x) = bin_width * ( bin_number(x) + 0.5 )

file=ARG1

plot file using (rounded($1)):(1) smooth frequency with boxes
pause -1