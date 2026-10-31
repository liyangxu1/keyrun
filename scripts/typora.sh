#!/bin/bash
# Typora macOS 工具：使用系统 Bash、Perl 和 OpenSSL，无需 Node。
# 参考：https://github.com/FlyAIBox/typora-activate-for-mac/blob/main/mac.js
set -euo pipefail
umask 077

usage() {
    cat <<'EOF'
用法：bash typora.sh [--app /Applications/Typora.app] [--restore]
      bash typora.sh --help

执行前请保存文档并退出 Typora。脚本会备份完整应用，再替换公钥、
生成离线 Token 并重新签名。仅适用于包含原始公钥的 macOS 版本。
--restore 从应用旁的 .keyrun-backup 目录恢复应用；不删除许可证偏好设置。
不需要 Node、npm、Python 或 Homebrew，不会自动安装依赖。
EOF
}

fail() { printf '错误：%s\n' "$*" >&2; exit 1; }

# 二进制处理和 JSON 编码交给系统 Perl，避免 Shell 丢失空字节。
process_data() {
    /usr/bin/perl -MJSON::PP -MMIME::Base64 - "$@" <<'PERL'
use strict;
use warnings;
sub read_bytes {
    my ($path) = @_;
    open my $fh, '<:raw', $path or die "读取失败：$!\n";
    local $/;
    return <$fh>;
}
sub write_bytes {
    my ($path, $data) = @_;
    open my $fh, '>:raw', $path or die "写入失败：$!\n";
    print {$fh} $data or die "写入失败：$!\n";
    close $fh or die "关闭文件失败：$!\n";
}
my ($mode, $dir) = @ARGV;
if ($mode eq 'payload') {
    my $code = read_bytes("$dir/machine");
    $code =~ s/\s+//g;
    die "机器码不是有效的 Base64\n" unless length($code) && $code =~ m{\A[A-Za-z0-9+/]+={0,2}\z} && length($code) % 4 == 0;
    my $mc = JSON::PP->new->utf8->decode(decode_base64($code));
    die "机器码必须是 JSON 对象\n" unless ref($mc) eq 'HASH';
    for my $key (qw(l i v)) {
        die "机器码缺少有效字段：$key\n" unless defined($mc->{$key}) && !ref($mc->{$key}) && length($mc->{$key});
    }
    my $email = read_bytes("$dir/email");
    # 将终端的 UTF-8 字节解码，JSON 序列化时再统一编码。
    require Encode;
    $email = Encode::decode('UTF-8', $email, Encode::FB_CROAK());
    my $date = read_bytes("$dir/date");
    my $entity = {deviceId => $mc->{l}, fingerprint => $mc->{i}, version => $mc->{v},
        email => $email, date => $date, license => 'Cracked_By_DreamNya&Letr', type => 'DreamNya'};
    my $json = JSON::PP->new->utf8->canonical->encode($entity);
    write_bytes("$dir/payload", encode_base64($json, ''));
    # 与原脚本一致：浏览器 atob 将 UTF-8 字节当作 Latin1 字符。
    my $decoded = JSON::PP->new->decode(Encode::decode('ISO-8859-1', $json));
    my $message = join '', map { $decoded->{$_} } sort keys %$decoded;
    write_bytes("$dir/message", Encode::encode('UTF-8', $message));
} elsif ($mode eq 'patch') {
    my ($source, $dest) = @ARGV[2,3];
    my $original = <<'KEY';
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7nVoGCHqIMJyqgALEUrc
5JJhap0+HtJqzPE04pz4y+nrOmY7/12f3HvZyyoRsxKdXTZbO0wEHFIh0cRqsuaJ
PyaOOPbA0BsalofIAY3mRhQQ3vSf+rn3g+w0S+udWmKV9DnmJlpWqizFajU4T/E4
5ZgMNcXt3E1ips32rdbTR0Nnen9PVITvrbJ3l6CI2BFBImZQZ2P8N+LsqfJsqyVV
wDkt3mHAVxV7FZbfYWG+8FDSuKQHaCmvgAtChx9hwl3J6RekkqDVa6GIV13D23LS
qdk0Jb521wFJi/V6QAK6SLBiby5gYN6zQQ5RQpjXtR53MwzTdiAzGEuKdOtrY2Me
DwIDAQAB
-----END PUBLIC KEY-----
KEY
    chomp $original;
    my $key = read_bytes("$dir/public.pem");
    $key =~ s/\s+\z//;
    die "公钥长度不匹配\n" unless length($key) == length($original);
    my $binary = read_bytes($source);
    my $count = ($binary =~ s/\Q$original\E/$key/g);
    die "未找到原始公钥，版本不兼容或已被修改；未修改应用\n" unless $count;
    for my $url ('https://store.typora.io', 'https://dian.typora.com.cn') {
        my $replacement = 'https://127.0.0.1' . ("\0" x (length($url) - length('https://127.0.0.1')));
        $binary =~ s/\Q$url\E/$replacement/g;
    }
    write_bytes($dest, $binary);
    print "已准备 $count 处公钥替换。\n";
} else { die "未知处理模式\n"; }
PERL
}

main() {
    local app='/Applications/Typora.app' restore=0 machine email binary backup command
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --help|-h) usage; return ;;
            --app) [[ $# -ge 2 && -n "$2" ]] || fail '--app 缺少路径'; app="$2"; shift 2 ;;
            --restore) restore=1; shift ;;
            *) fail "未知参数：$1" ;;
        esac
    done
    [[ $(uname -s) == Darwin ]] || fail '仅支持 macOS'
    [[ $EUID -ne 0 ]] || fail '请使用当前登录用户执行，不要使用 sudo'
    for command in perl openssl codesign ditto pgrep; do
        [[ -x "/usr/bin/$command" ]] || fail "缺少系统工具：$command；请手动处理，不会自动安装"
    done
    /usr/bin/perl -MJSON::PP -MMIME::Base64 -MEncode -e 1 || fail '系统 Perl 模块不可用'
    app="${app%/}"
    [[ -d "$app" && ! -L "$app" ]] || fail "应用目录不存在或是符号链接：$app"
    app="$(cd "$(dirname "$app")" && pwd -P)/$(basename "$app")"
    binary="$app/Contents/MacOS/Typora"
    backup="${app}.keyrun-backup"
    [[ -f "$binary" && ! -L "$binary" ]] || fail '未找到正常的 Typora 二进制文件'
    if /usr/bin/pgrep -x Typora >/dev/null; then fail '请保存文档并退出 Typora 后重试'; fi
    [[ -w "$binary" && -w "$(dirname "$app")" ]] || fail '应用或父目录不可写，请将应用放到当前用户可写的位置'
    if [[ $restore == 1 ]]; then
        [[ -f "$backup/Contents/MacOS/Typora" && ! -L "$backup" ]] || fail '未找到完整应用备份'
        /usr/bin/ditto "$backup" "$app"
        printf '已恢复应用；备份保留在：%s\n' "$backup"
        return
    fi
    [[ ! -e "$backup" ]] || fail "已有备份：$backup。可用 --restore 恢复；重新处理前请先另行归档备份"
    [[ -r /dev/tty ]] || fail '需要在交互终端输入机器码和邮箱'
    printf '请输入机器码：' >/dev/tty
    IFS= read -r machine </dev/tty || fail '未读取到机器码'
    printf '请输入邮箱：' >/dev/tty
    IFS= read -r email </dev/tty || fail '未读取到邮箱'
    [[ -n "$machine" && -n "$email" ]] || fail '机器码和邮箱不能为空'
    work_dir=$(mktemp -d "${TMPDIR:-/tmp}/keyrun-typora.XXXXXX")
    modified=0
    # 异常退出时恢复应用，临时私钥和机器信息不留存。
    trap 'result=$?; trap - EXIT; if [[ $modified == 1 ]]; then /usr/bin/ditto "$backup" "$app" || printf "自动恢复失败，请使用 --restore。\n" >&2; fi; rm -rf "$work_dir"; exit "$result"' EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    printf '%s' "$machine" > "$work_dir/machine"
    printf '%s' "$email" > "$work_dir/email"
    date '+%m/%d/%Y' | tr -d '\n' > "$work_dir/date"
    process_data payload "$work_dir"
    /usr/bin/openssl genrsa -out "$work_dir/private.pem" 2048 2>/dev/null
    /usr/bin/openssl rsa -in "$work_dir/private.pem" -pubout -out "$work_dir/public.pem" 2>/dev/null
    process_data patch "$work_dir" "$binary" "$work_dir/patched"
    /usr/bin/openssl dgst -sha256 -sign "$work_dir/private.pem" -out "$work_dir/signature" "$work_dir/message"
    { printf '+'; cat "$work_dir/payload"; printf '|'; /usr/bin/openssl base64 -A -in "$work_dir/signature"; printf '#'; } > "$work_dir/token"
    /usr/bin/ditto "$app" "$backup"
    modified=1
    cat "$work_dir/patched" > "$binary"
    /usr/bin/codesign --force --deep -s - "$app"
    /usr/bin/codesign --verify --deep "$app"
    modified=0
    printf '\n处理完成。打开 Typora → 帮助 → 我的许可证 → 离线激活，粘贴以下 Token：\n\n'
    cat "$work_dir/token"
    printf '\n\n'
    if /usr/bin/pbcopy < "$work_dir/token"; then printf 'Token 已复制到剪贴板。\n'; fi
    printf '应用备份：%s\n' "$backup"
    rm -rf "$work_dir"
    trap - EXIT INT TERM
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
