; CS2 Better Auto Director Installer Script for Inno Setup
; Compile with Inno Setup: https://jrsoftware.org/isinfo.php

[Setup]
AppName=CS2 Better Auto Director
AppVersion=0.2.3
AppPublisher=CS2 Better Auto Director
DefaultDirName={autopf}\CS2BetterAutoDirector
DefaultGroupName=CS2 Better Auto Director
OutputDir=..\build
OutputBaseFilename=CS2BetterAutoDirector-Setup
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequiredOverridesAllowed=dialog
UninstallDisplayIcon={app}\cs2-better-autodirector.exe
ShowLanguageDialog=no

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "german"; MessagesFile: "compiler:Languages\German.isl"

[Files]
; Main GUI program
Source: "..\gui\build\bin\cs2-better-autodirector.exe"; DestDir: "{app}"; Flags: ignoreversion

; Config files
Source: "..\config\gamestate_integration_autodirector.cfg"; DestDir: "{app}\config"; Flags: ignoreversion
Source: "..\config\spectator_bindings.cfg"; DestDir: "{app}\config"; Flags: ignoreversion

[Icons]
Name: "{group}\CS2 Better Auto Director"; Filename: "{app}\cs2-better-autodirector.exe"; WorkingDir: "{app}"
Name: "{autoprograms}\CS2 Better Auto Director"; Filename: "{app}\cs2-better-autodirector.exe"; WorkingDir: "{app}"
Name: "{autodesktop}\CS2 Better Auto Director"; Filename: "{app}\cs2-better-autodirector.exe"; WorkingDir: "{app}"; Tasks: desktopicon
Name: "{group}\Config Folder"; Filename: "{app}\config"
Name: "{group}\GitHub Repository"; Filename: "https://github.com/zGLados/cs2-better-autodirector"
Name: "{group}\Uninstall CS2 Better Auto Director"; Filename: "{uninstallexe}"

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop icon"; GroupDescription: "Additional icons:"
Name: "copygsiconfig"; Description: "Copy Game State Integration config to CS2 cfg folder"; GroupDescription: "CS2 Integration:"

[Run]
Filename: "https://github.com/zGLados/cs2-better-autodirector"; Description: "Open GitHub Repository"; Flags: postinstall shellexec skipifsilent unchecked
Filename: "{app}\cs2-better-autodirector.exe"; Description: "Launch CS2 Better Auto Director"; Flags: postinstall nowait skipifsilent; Check: not IsAdminInstallMode

[UninstallDelete]
Type: filesandordirs; Name: "{app}"

[Code]
var
  CS2ConfigPage: TInputDirWizardPage;
  CS2ConfigPath: String;

function GetCS2ConfigPath(): String;
var
  SteamPath: String;
begin
  Result := '';
  
  // Try to get Steam path from registry
  if RegQueryStringValue(HKEY_LOCAL_MACHINE, 'SOFTWARE\WOW6432Node\Valve\Steam', 'InstallPath', SteamPath) or
     RegQueryStringValue(HKEY_LOCAL_MACHINE, 'SOFTWARE\Valve\Steam', 'InstallPath', SteamPath) then
  begin
    // Check default Steam library
    if DirExists(SteamPath + '\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    begin
      Result := SteamPath + '\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg';
      Exit;
    end;
  end;
  
  // Try standard paths
  if DirExists('C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg'
  else if DirExists('D:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'D:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg'
  else if DirExists('E:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'E:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg'
  else if DirExists('F:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg') then
    Result := 'F:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg';
end;

procedure InitializeWizard();
var
  DetectedPath: String;
begin
  // Create the CS2 config path input page
  CS2ConfigPage := CreateInputDirPage(wpSelectTasks,
    'CS2 Config Folder Location', 
    'Where should the Game State Integration config be copied?',
    'Setup has detected your CS2 config folder. If the path is incorrect, you can change it below.' + #13#10 + #13#10 +
    'The config folder is typically located at:' + #13#10 +
    'Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg',
    False, '');
  
  // Add the directory input field FIRST
  CS2ConfigPage.Add('');
  
  // Get auto-detected path
  DetectedPath := GetCS2ConfigPath();
  
  // Set the value AFTER adding the field
  if DetectedPath <> '' then
    CS2ConfigPage.Values[0] := DetectedPath
  else
    CS2ConfigPage.Values[0] := 'C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg';
end;

function ShouldSkipPage(PageID: Integer): Boolean;
begin
  Result := False;
  
  // Skip CS2 config page if user unchecked the task
  if PageID = CS2ConfigPage.ID then
    Result := not WizardIsTaskSelected('copygsiconfig');
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  SourceFile: String;
  DestFile: String;
begin
  if CurStep = ssPostInstall then
  begin
    // Only copy if task is selected
    if WizardIsTaskSelected('copygsiconfig') then
    begin
      CS2ConfigPath := CS2ConfigPage.Values[0];
      
      if DirExists(CS2ConfigPath) then
      begin
        SourceFile := ExpandConstant('{app}\config\gamestate_integration_autodirector.cfg');
        DestFile := CS2ConfigPath + '\gamestate_integration_autodirector.cfg';
        
        // Copy config file silently (no success popup)
        if not CopyFile(SourceFile, DestFile, False) then
        begin
          MsgBox('Failed to copy config file to:' + #13#10 + 
                 CS2ConfigPath + #13#10 + #13#10 +
                 'Please copy manually from:' + #13#10 +
                 ExpandConstant('{app}\config\gamestate_integration_autodirector.cfg'),
                 mbError, MB_OK);
        end;
      end
      else
      begin
        MsgBox('CS2 config folder not found:' + #13#10 + 
               CS2ConfigPath + #13#10 + #13#10 +
               'Please manually copy:' + #13#10 +
               ExpandConstant('{app}\config\gamestate_integration_autodirector.cfg') + #13#10 + #13#10 +
               'To your CS2 config folder.',
               mbError, MB_OK);
      end;
    end;
  end;
  
  // Set admin flag AFTER postinstall to avoid startup issues
  if CurStep = ssDone then
  begin
    if IsAdminInstallMode then
    begin
      // Set Start Menu shortcut to run as admin
      RegWriteStringValue(HKLM, 
        'SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\cs2-better-autodirector.exe',
        '', ExpandConstant('{app}\cs2-better-autodirector.exe'));
      
      // Set compatibility layer for the exe to always run as admin
      RegWriteStringValue(HKLM,
        'SOFTWARE\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers',
        ExpandConstant('{app}\cs2-better-autodirector.exe'),
        'RUNASADMIN');
    end;
  end;
end;
