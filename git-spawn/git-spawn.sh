#!/bin/bash

usage="Usage: $0 -i {old-repo} -o {new-repo} [-f {config-file}] [-h]

This script is used to create a new git repository from an existing
one.  In the process of creating the new repository, it calls
git-filter-repo to shape the layout of the new repository, moving or
removing file or directories.  The actions to be taken are specified
in a file.  The file is used as input to the git-filter-repo
--paths-from-file option.  If the file is not specified, the script
generates an example file in /tmp and opens it in the default editor.
"

example_config='
# Keep the file or directory baz
# equivalent to --path baz
# literal:baz

# Keep all files or directories matching the pattern foo/*
# glob:foo/*

# Keep all files or directories matching the regex pattern ^foo/
# regex:^foo/

# Move baz directory contents to the root of the repository
# equivalent to --path-rename baz/:
# (use no spaces around the arrow)
# literal:baz/==>

# git filter-repo --analyze found the following renames:
'

while getopts "i:o:f:h" opt; do
    case $opt in
        i)
            old_repo_dir=$OPTARG
            ;;
        o)
            new_repo_dir=$OPTARG
            ;;
        f)
            config_file=$OPTARG
            ;;
        h)
            echo "$usage"
            exit 0
            ;;
        \?)
            echo "Invalid option: $OPTARG" 1>&2
            echo "$usage" 1>&2
            exit 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." 1>&2
            echo "$usage" 1>&2
            exit 1
            ;;
    esac
done

create_config() {
    config_file=/tmp/git-spawn-$$.txt
    echo config file is $config_file

    # Create an example config file
    echo "$example_config" > $config_file

    # Append the renames found by git filter-repo --analyze
    git filter-repo --analyze
    perl -pne 's/^/# /' .git/filter-repo/analysis/renames.txt >> $config_file

    # Open the config file in the default editor
    ${EDITOR:-vi} $config_file
}

main() {
    if [ -z "$new_repo_dir" ]; then
        echo "Option -o is required." 1>&2
        echo "$usage" 1>&2
        exit 1
    fi

    if [ -e "$new_repo_dir" ]; then
        echo "Directory $new_repo_dir already exists." 1>&2
        exit 1
    fi

    set -ex

    git clone --no-local $old_repo_dir $new_repo_dir
    cd $new_repo_dir

    if [ -z "$config_file" ]; then
        create_config
    fi

    echo "Enter to continue, ^C to abort"
    read

    git filter-repo --paths-from-file $config_file
}

main
