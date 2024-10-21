---
title: Adding Assembly (asm) Files to Visual Studio C/C++ Projects
---

### Step 0
Create a new empty C/C++ project and add some C/C++ source code or open an existing project.
![Start](/assets/img/2024-10-21/start.png)

### Step 1
Go `Project -> Build Customization...`
![Step 1](/assets/img/2024-10-21/step1.png)

### Step 2
In the `Visual C++ Build Customization Files` dialog, tick `masm(.targets, .props)` and click on `OK`.
![Step 2](/assets/img/2024-10-21/step2.png)

### Step 3
Right click `Source Files`, choose `Add`, then `New Item...`
![Step 3](/assets/img/2024-10-21/step3.png)

### Step 4
In the `Add New Item` dialog, select `C++ File (.cpp)`, and name the file with a `.asm` extension, then click `Add`.
**NOTE: that the `.asm` extension is important.**
![Step 4](/assets/img/2024-10-21/step4.png)

### Step 5
Write some asm code.
![Step 5](/assets/img/2024-10-21/step5.png)

### Step 6
Update `main.c` to call the new asm function.
![Step 6](/assets/img/2024-10-21/step6.png)

### Step 7
Everything should work.
![Step 7](/assets/img/2024-10-21/step7.png)

### Step 8
If you have issue, check the `.asm` file by right clicking, and going `Properties`.
![Step 8](/assets/img/2024-10-21/step8.png)

### Step 9
In the dialog, ensure `Item Type` is `Microsoft Macro Assembler` then click `OK`.
![Step 9](/assets/img/2024-10-21/step9.png)
