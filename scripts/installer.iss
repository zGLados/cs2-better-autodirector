; Better Auto Observer Installer Script for Inno Setup
; Compile with Inno Setup: https://jrsoftware.org/isinfo.php

[Setup]
AppName=Better Auto Observer
AppVersion=1.0
AppPublisher=Better Auto Observer
DefaultDirName={autopf}\BetterAutoObserver
DefaultGroupName=Better Auto Observer
OutputBaseFilename=BetterAutoObserver-Setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
SetupIconFile=

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "german"; MessagesFile: "compiler:Languages\German.isl"

[Files]
; Main program
Source: "better-autoobserver.exe"; DestDir: "{app}"; Flags: ignoreversion

; GSI Config
Source: "gamestate_integration_autoobserver.cfg"; DestDir: "{app}"; Flags: ignoreversion

; README
Source: "README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Better Auto Observer"; Filename: "{app}\better-autoobserver.exe"
Name: "{group}\README"; Filename: "{app}\README.md"
Name: "{group}\Uninstall"; Filename: "{uninstallexe}"

[Run]
; Show README after installation
Filename: "{app}\README.md"; Description: "View README"; Flags: postinstall shellexec skipifsilent

[Code]
var
  CSConfigPath: String;
  ConfigCopied: Boolean;

function GetCSConfigPath(): String;
var
  SteamPath: String;
begin
  Result := '';
  
  // Try standard Steam paths
  if DirExists('C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg'
  else if DirExists('D:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'D:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg'
  else if DirExists('E:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'E:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg';
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
  SourceFile: String;
  DestFile: String;
begin
  if CurStep = ssPostInstall then
  begin
    ConfigCopied := False;
    CSConfigPath := GetCSConfigPath();
    
    if CSConfigPath <> '' then
    begin
      SourceFile := ExpandConstant('{app}\gamestate_integration_autoobserver.cfg');
      DestFile := CSConfigPath + '\gamestate_integration_autoobserver.cfg';
      
      if FileCopy(SourceFile, DestFile, False) then
      begin
        ConfigCopied := True;
        MsgBox('Game State Integration config successfully copied to:' + #13#10 + 
               CSConfigPath, mbInformation, MB_OK);
      end;
    end;
    
    if not ConfigCopied then
    begin
      MsgBox('Could not automatically copy the GSI config.' + #13#10 + #13#10 +
             'Please manually copy:' + #13#10 +
             ExpandConstant('{app}\gamestate_integration_autoobserver.cfg') + #13#10 + #13#10 +
             'To your CS config folder:' + #13#10 +
             'Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\',
             mbInformation, MB_OK);
    end;
  end;
end;

procedure InitializeWizard();
begin
  // Custom welcome page could be added here
end;
