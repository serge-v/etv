echo startup file loaded
source /home/alarm/params.sh
omxplayer --win '0 0 320 240' --dbus_name org.mpris.MediaPlayer2.omxplayer2 ${URL}1.mp4 &
