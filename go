#!/bin/bash

# set -x

usage() {
	echo "$0 <type> <arch>" >&2
	echo >&2
	echo "i.e.," >&2
	echo "# $0 fedora-30 x86_64" >&2
	echo >&2
	echo "nb: First argument must contain hyphen followed by release number" >&2
	echo "    Neither argument may contain whitespace" >&2
	echo "    Only supported types: fedora-29 fedora-3[012] rhel-[78]"
	echo "    only x86_64 supported on rhel-[78]"
	exit 1
}

if [[ $# -ne 2 ]]; then
	usage
fi

for arg in "$@"; do
	case arg in
		[-/][hH?]|--help|[/-][hH][eE][lL][pP])
			usage
			;;
	esac
done

[[ ${1}${2} == *\ * ]] && usage
[[ ${1} == *-* ]] || usage

btype=${1%-*}
brel=${1##*-}
btypedrel=${1}
barch=${2}
bmocktype="${btypedrel}-${barch}"

[[ -f "./etc_mock_default.cfg_${btypedrel}" ]] || {
	echo "File etc_mock_default.cfg_${btypedrel} not present in ${PWD}" >&2
	exit 1
}

cp "./etc_mock_default.cfg_${btypedrel}"  /etc/mock/default.cfg || {
	echo "Failed copying to /etc/mock/default.cfg" >&2
	exit 1
}

[[ -d /share/yumrepo/${btypedrel^} ]] || {
	echo "Directory /share/yumrepo/${btypedrel^} dne, not mounted?" >&2
	exit 1
}

setenforce 0
[[ $(getenforce) == Disabled ]] \
	|| { echo "Cant disable selinux" >&2; exit 1; }

mock -r ${bmocktype} --clean --init \
	|| { echo "mock init invocation failed" >&2; exit 1; }

mock -r ${bmocktype} --copyin /share/yumrepo /share/yumrepo \
	|| { echo "mock /share/yumrepo copyin failed" >&2; exit 1; }

mock -r ${bmocktype} --install git lorax-lmc-novirt vim-minimal pykickstart \
	libblockdev-{lvm,swap,loop,crypto,mpath,btrfs,dm,mdraid,nvdimm} \
		|| { echo "mock install failed" >&2; exit 1; }

mock -r ${bmocktype} --old-chroot --shell \
	"cd /root && git clone https://pagure.io/fedora-kickstarts" \
	|| { echo "git clone failed" >&2; exit 1; }

myks=bsksdne

case ${btype} in
	fedora)
		myversion="F${brel}"
		myks=fedora-live-workstation.ks
		;;
	*)
		echo OMGWTFBBQ >&2
		exit 1
		;;
esac

mock -r ${bmocktype} --old-chroot --shell "$(echo cd /root/fedora-kickstarts \&\& \
ksflatten --config ${myks} -o flat-${myks} --version ${myversion} )" || {
	echo "flatten failed" >&2
	exit 1
}

case ${btype} in
	fedora)
		myproject=Fedora-workstation-Live
		mytitle=Fedora-Workstation-live
		myvolid="FWKLV${brel}"
		myisoname="${btype^}-Workstation-Live-${brel}-ntfswip-${barch}.iso"
		;;
	*)
		# TODO: rhel
		echo OASEDFASDF >&2
		exit 1
		;;
esac

mock -r ${bmocktype} --old-chroot --shell "cd /root/fedora-kickstarts && livemedia-creator \
	--ks flat-${myks} --no-virt --resultdir /var/lmc --project ${myproject} --make-iso \
	--volid ${myvolid} --iso-only --iso-name ${myisoname} --releasever ${brel} \
	--title ${mytitle} --macboot" || {
		echo "mock -> livemedia-creator failed." >&2
		exit 1
	}

mock -r ${bmocktype} --copyout /var/lmc/${myisoname} ./iso/${myisoname} || {
	echo "failed to extract iso from chroot" >&2
	exit 1
}

echo "Success?"

exit 0
