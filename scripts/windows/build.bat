
IF EXIST "bin" (
  echo -ne 'Building... [****              ](25%)\r'
  sleep 1s
  echo -ne 'Building... [********          ](50%)\r'
  sleep 1s
  echo -ne 'Building... [***********       ](75%)\r'
  sleep 1s
  echo -ne 'Building... [***************** ](99%)\r'
  sleep 1s
  mv -f main bin/
  echo -ne 'Building... [******************](100%)\r'
) ELSE (
  mkdir bin
  echo -ne 'Building... [****              ](25%)\r'
  sleep 1s
  echo -ne 'Building... [********          ](50%)\r'
  sleep 1s
  echo -ne 'Building... [***********       ](75%)\r'
  sleep 1s
  echo -ne 'Building... [***************** ](99%)\r'
  sleep 1s
  mv -f main bin/
  echo -ne 'Building... [******************](100%)\r'
)