mkisofs notes


mkisofs \
   -o boot.iso \
     # -o = name of output .iso file
   
   -R -J -v -d -N \
     # -R = enables normal Unix filenames and attributes by Rock Ridge extension.
     # -J = output's ISO using Joliet format (useful for Windows users of the final ISO)
     # -v = ?
     # -d = ?
     # -N = ?

   -hide-rr-moved \
     # -hide-rr-moved = hides the directory RR_MOVED to .rr_moved

   -no-emul-boot \
     # -no-emul-boot = choose emulation modes, is needed for boot images of ISOLINUX and GRUB2.

   -eltorito-platform=efi \
   
   -V "EFIBOOTISO" \
     # -V = Volume ID

   -A "EFI Boot ISO" \
     # -A = application_id, a text string written into the volume header.
        
   root




# -r = set permissions to 0