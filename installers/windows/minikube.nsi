# This installs two files, minikube.exe and logo.ico, creates a start menu shortcut, builds an uninstaller, and
# adds uninstall information to the registry for Add/Remove Programs

# To get started, put this script into a folder with the two files (minikube.exe, logo.ico, and LICENSE.txt -
# You'll have to create these yourself) and run makensis on it

# If you change the names "minikube.exe", "logo.ico", or "LICENSE.txt" you should do a search and replace - they
# show up in a few places.
# All the other settings can be tweaked by editing the !defines at the top of this script
# Unicode true # This command is not available in 2.46 which is on apt-get. Debating how to proceed
!define APPNAME "Minikube"
!define COMPANYNAME "Kubernetes"
!define DESCRIPTION "A Local Kubernetes Development Environment"
# These three must be integers
!define VERSIONMAJOR --VERSION_MAJOR--
!define VERSIONMINOR --VERSION_MINOR--
!define VERSIONBUILD --VERSION_BUILD--
# These will be displayed by the "Click here for support information" link in "Add/Remove Programs"
# It is possible to use "mailto:" links in here to open the email client
!define HELPURL "https://github.com/kubernetes/minikube" # "Support Information" link
!define UPDATEURL "https://github.com/kubernetes/minikube/releases" # "Product Updates" link
!define ABOUTURL "https://github.com/kubernetes/minikube" # "Publisher" link
# This is the size (in kB) of all the files copied into "Program Files"
!define INSTALLSIZE --INSTALL_SIZE--

RequestExecutionLevel admin ;Require admin rights on NT6+ (When UAC is turned on)

InstallDir "$PROGRAMFILES64\${COMPANYNAME}\${APPNAME}"
!define UNINSTALLDIR "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}"
BrandingText " "

# rtf or txt file - remember if it is txt, it must be in the DOS text format (\r\n)
# This will be in the installer/uninstaller's title bar
Name "${COMPANYNAME} - ${APPNAME}"
Icon "logo.ico"
OutFile "minikube-installer.exe"

!include "LogicLib.nsh"
!include "MUI2.nsh"       ; Modern UI

!define MUI_ICON "logo.ico"
!define MUI_UNICON "logo.ico"
!define MUI_WELCOMEFINISHPAGE_BITMAP "logo.bmp"
!define MUI_UNWELCOMEFINISHPAGE_BITMAP "logo.bmp"
!define MUI_HEADERIMAGE_BITMAP "logo.bmp"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Set languages (first is default language)
;!insertmacro MUI_LANGUAGE "English"
!define MUI_LANGDLL_ALLLANGUAGES
;Languages

  !insertmacro MUI_LANGUAGE "English"
  !insertmacro MUI_LANGUAGE "French"
  !insertmacro MUI_LANGUAGE "TradChinese"
  !insertmacro MUI_LANGUAGE "Spanish"
  !insertmacro MUI_LANGUAGE "Hungarian"
  !insertmacro MUI_LANGUAGE "Russian"
  !insertmacro MUI_LANGUAGE "German"
  !insertmacro MUI_LANGUAGE "Dutch"
  !insertmacro MUI_LANGUAGE "SimpChinese"
  !insertmacro MUI_LANGUAGE "Italian"
  !insertmacro MUI_LANGUAGE "Danish"
  !insertmacro MUI_LANGUAGE "Polish"
  !insertmacro MUI_LANGUAGE "Czech"
  !insertmacro MUI_LANGUAGE "Slovenian"
  !insertmacro MUI_LANGUAGE "Slovak"
  !insertmacro MUI_LANGUAGE "Swedish"
  !insertmacro MUI_LANGUAGE "Norwegian"
  !insertmacro MUI_LANGUAGE "PortugueseBR"
  !insertmacro MUI_LANGUAGE "Ukrainian"
  !insertmacro MUI_LANGUAGE "Turkish"
  !insertmacro MUI_LANGUAGE "Catalan"
  !insertmacro MUI_LANGUAGE "Arabic"
  !insertmacro MUI_LANGUAGE "Lithuanian"
  !insertmacro MUI_LANGUAGE "Finnish"
  !insertmacro MUI_LANGUAGE "Greek"
  !insertmacro MUI_LANGUAGE "Korean"
  !insertmacro MUI_LANGUAGE "Hebrew"
  !insertmacro MUI_LANGUAGE "Portuguese"
  !insertmacro MUI_LANGUAGE "Farsi"
  !insertmacro MUI_LANGUAGE "Bulgarian"
  !insertmacro MUI_LANGUAGE "Indonesian"
  !insertmacro MUI_LANGUAGE "Japanese"
  !insertmacro MUI_LANGUAGE "Croatian"
  !insertmacro MUI_LANGUAGE "Serbian"
  !insertmacro MUI_LANGUAGE "Thai"
  !insertmacro MUI_LANGUAGE "NorwegianNynorsk"
  !insertmacro MUI_LANGUAGE "Belarusian"
  !insertmacro MUI_LANGUAGE "Albanian"
  !insertmacro MUI_LANGUAGE "Malay"
  !insertmacro MUI_LANGUAGE "Galician"
  !insertmacro MUI_LANGUAGE "Basque"
  !insertmacro MUI_LANGUAGE "Luxembourgish"
  !insertmacro MUI_LANGUAGE "Afrikaans"
  !insertmacro MUI_LANGUAGE "Uzbek"
  !insertmacro MUI_LANGUAGE "Macedonian"
  !insertmacro MUI_LANGUAGE "Latvian"
  !insertmacro MUI_LANGUAGE "Bosnian"
  !insertmacro MUI_LANGUAGE "Mongolian"
  !insertmacro MUI_LANGUAGE "Estonian"

!insertmacro MUI_RESERVEFILE_LANGDLL

Function .onInit

  !insertmacro MUI_LANGDLL_DISPLAY

FunctionEnd

Section "Install"
	# Files for the install directory - to build the installer, these should be in the same directory as the install script (this file)
	SetOutPath $INSTDIR
	# Files added here should be removed by the uninstaller (see section "uninstall")
	File "minikube.exe"
	File "logo.ico"
    File "update_path.bat"
	# Add any other files for the install directory (license files, app data, etc) here

	# Uninstaller - See function un.onInit and section "uninstall" for configuration
	WriteUninstaller "$INSTDIR\uninstall.exe"

	# Start Menu
	CreateDirectory "$SMPROGRAMS\${COMPANYNAME}"
	CreateShortCut "$SMPROGRAMS\${COMPANYNAME}\${APPNAME}.lnk" "$INSTDIR\minikube.exe" "" "$INSTDIR\logo.ico"

	# Registry information for add/remove programs
	WriteRegStr HKLM "${UNINSTALLDIR}" "DisplayName" "${COMPANYNAME} - ${APPNAME} - ${DESCRIPTION}"
	WriteRegStr HKLM "${UNINSTALLDIR}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
	WriteRegStr HKLM "${UNINSTALLDIR}" "InstallLocation" "$\"$INSTDIR$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "DisplayIcon" "$\"$INSTDIR\logo.ico$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "Publisher" "$\"${COMPANYNAME}$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "HelpLink" "$\"${HELPURL}$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "URLUpdateInfo" "$\"${UPDATEURL}$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "URLInfoAbout" "$\"${ABOUTURL}$\""
	WriteRegStr HKLM "${UNINSTALLDIR}" "DisplayVersion" "$\"${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}$\""
	WriteRegDWORD HKLM "${UNINSTALLDIR}" "VersionMajor" ${VERSIONMAJOR}
	WriteRegDWORD HKLM "${UNINSTALLDIR}" "VersionMinor" ${VERSIONMINOR}
	# There is no option for modifying or repairing the install
	WriteRegDWORD HKLM "${UNINSTALLDIR}" "NoModify" 1
	WriteRegDWORD HKLM "${UNINSTALLDIR}" "NoRepair" 1
	# Set the INSTALLSIZE constant (!defined at the top of this script) so Add/Remove Programs can accurately report the size
	WriteRegDWORD HKLM "${UNINSTALLDIR}" "EstimatedSize" ${INSTALLSIZE}

	# Add installed executable to PATH
    # Cannot uset EnvVarUpdate since the path can be too long
    # this is explicitly warned in the documentation page
    # http://nsis.sourceforge.net/Environmental_Variables:_append,_prepend,_and_remove_entries
    nsExec::Exec '"$INSTDIR\update_path.bat" add $INSTDIR'
SectionEnd

Section "Uninstall"

	# Remove Start Menu launcher
	Delete /REBOOTOK "$SMPROGRAMS\${COMPANYNAME}\${APPNAME}.lnk"
	# Try to remove the Start Menu folder - this will only happen if it is empty
	RmDir /REBOOTOK "$SMPROGRAMS\${COMPANYNAME}"

	# Remove uninstalled executable from PATH
    nsExec::Exec '"$INSTDIR\update_path.bat" remove $INSTDIR' ; appends to the system path

	# Remove files
	Delete /REBOOTOK $INSTDIR\minikube.exe
	Delete /REBOOTOK $INSTDIR\logo.ico
	Delete /REBOOTOK $INSTDIR\update_path.bat

	# Always delete uninstaller as the last action
	Delete /REBOOTOK $INSTDIR\uninstall.exe

	# Try to remove the install directory - this will only happen if it is empty
	RmDir /REBOOTOK $INSTDIR

	# Remove uninstaller information from the registry
	DeleteRegKey HKLM "${UNINSTALLDIR}"

SectionEnd
