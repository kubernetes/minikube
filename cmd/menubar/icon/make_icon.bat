@ECHO OFF

IF "%GOPATH%"=="" GOTO NOGO
IF NOT EXIST %GOPATH%\bin\2goarray.exe GOTO INSTALL
:POSTINSTALL
IF "%1"=="" GOTO NOICO
IF NOT EXIST %1 GOTO BADFILE
ECHO Creating iconwin.go
ECHO //+build windows > iconwin.go
ECHO. >> iconwin.go
TYPE %1 | %GOPATH%\bin\2goarray Data icon >> iconwin.go
GOTO DONE

:CREATEFAIL
ECHO Unable to create output file
GOTO DONE

:INSTALL
ECHO Installing 2goarray...
go get github.com/cratonica/2goarray
IF ERRORLEVEL 1 GOTO GETFAIL
GOTO POSTINSTALL

:GETFAIL
ECHO Failure running go get github.com/cratonica/2goarray.  Ensure that go and git are in PATH
GOTO DONE

:NOGO
ECHO GOPATH environment variable not set
GOTO DONE

:NOICO
ECHO Please specify a .ico file
GOTO DONE

:BADFILE
ECHO %1 is not a valid file
GOTO DONE

:DONE

