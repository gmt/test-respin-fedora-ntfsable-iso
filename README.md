This is a kludge to test the effects of adding dmsquash-live-ntfs to the lorax-provided default module list
of dracut modules on Fedora live spins.  To use it, build some sort of burner vm with all the (undocumented)
prerequisite packages installed, extract gmt_test_lorax_ntfs_yum_repo.tar.xz into the root of your burner
vm's filesystem, and then run one of:

  ./respin fedora-{29,30} {i386,x86_64}

for example

  ./respin fedora-30 x86_64

It should create a Fedora-Workstation-Live iso in ./iso with the ntfs dracut module enabled.

In my testing as of 20190901, all of the above create working ISO's with NTFS boot enabled,
which I am able to boot in x86_64 efi and bios modes, and also i386 bios mode.  I did actually
test i386 efi in a vm and although it eventually crashed, I think I later discovered that
was due to a VM misconfiguration issue, not any fundamental problem.

After ISO generation, my testing process was to manually unpack the
ISO onto an NTFS filesystem created with Rufus' (hidden-behind-magic-keystroke) UEFI:NTFS 
BIOS/EFI hybrid mode, accompanied by a handcrafted grub.cfg like:

```
# correct location on stick
set extract_path="/test/Fedora-Workstation-Live-30-ntfswip-x86_64"

# correct uuid of stick as detected by udev
set rootuuid="0F00-0BA7"

#####

export extract_path
export rootuuid

insmod part_msdos
insmod part_gpt
insmod ntfs
insmod ext2

set default="0"

function load_video {
if [ -n "$efi" ]; then   insmod efi_gop; fi
if [ -n "$efi" ]; then   insmod efi_uga; fi
  insmod video_bochs
  insmod video_cirrus
  insmod all_video
}

load_video
set gfxpayload=keep
insmod gzio

set timeout=20

menuentry 'Start Fedora-workstation-Live 30' --class fedora --class gnu-linux --class gnu --class os {
	linux ${extract_path}/images/pxeboot/vmlinuz root=live:UUID=${rootuuid}  rd.live.image rd.live.dir=${extract_path}/LiveOS quiet
	initrd ${extract_path}/images/pxeboot/initrd.img
}

submenu 'Troubleshooting -->' {
	menuentry 'Start Fedora-workstation-Live 30 in basic graphics mode' --class fedora --class gnu-linux --class gnu --class os {
		linux ${extract_path}/images/pxeboot/vmlinuz root=live:UUID=${rootuuid}  rd.live.image rd.live.dir=${extract_path}/LiveOS nomodeset quiet
		initrd ${extract_path}/images/pxeboot/initrd.img
	}
}
```

Notes
=====

I built my stick with an MS-DOS partition table.  I haven't tested any of this with GPT on the stick but suspect it would work.

My USB stick is UAS SATA and responds to scsi inquiries as non-removable.  So concievably this all breaks under
normal usb-storage.  I didn't even bother testing for this possibility, I think it's unlikely.

I also have not tested casper persistence on NTFS yet.  This is a more meaningful omission, as >4G persistent stores or
"WUBI" style just-trying-Linux-on-Windows-box are two of the major possible use-cases for this and they are also more
likely not working.  If these are broken I will probably find out soon enough and endeavor to fix them eventually.

Secure Boot is of course not working with any of this, nor will it unlesss and until this is upstream.  Even then,
I don't know enough about SB to say if it might work in any meaningfully helpful way or not.

-gmt

