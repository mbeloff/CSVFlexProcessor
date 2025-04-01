@echo off
echo Installing dependencies...
call npm install

echo Building executable...
call npm run build

echo Build complete! Executables are in the dist folder.
pause 