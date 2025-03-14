#!/bin/bash

# Release the next version of the cloudy project. This will 
# update the release version (minor), commit the version, 
# and update all the local cloudy dependencies.
_debug="yes"
set -e

debug() {
    if [[ _debug == "yes" ]]; then
        echo $1
    fi
}

# Updates the cloudy version in a repo
updatecloudy() {
    echo ""
    echo "Updating Cloudy version $nextVersion for $1"
    dir="../$1"
    pushd $dir &> /dev/null
    git pull
    go get "github.com/appliedres/cloudy@$nextVersion"
    go mod tidy

    popd &> /dev/null
}

# finds a version number give a version string and position
versions() {
    my_string=$1  
    my_array=($(echo $my_string | tr "." "\n"))

    a=${my_array[$2]}
    a=${a/v/}       # Remove the first v
    a=${a/-*/}      # Remove everything after dash

    echo $a
}

# Verifies that the argument is a number
mustBeNumber() {
    re='^[0-9]+$'
    if ! [[ $1 =~ $re ]] ; then
    echo "error: $1 is not a number" >&2; exit 1
    fi
}

# MAIN CODE STARTS HERE
# ------------------------------------------------------------------------------------
echo ""
echo "          _____           __    .__                      .___    "
echo "         /  _  \ _______ |  | __|  |    ____   __ __   __| _/    "
echo "        /  /_\  \\_  __  \|  |/ /|  |   /  _ \ |  |  \ / __ |     "
echo "       /    |    \|  | \/|    < |  |__(  <_> )|  |  // /_/ |     "
echo "       \____|__  /|__|   |__|_ \|____/ \____/ |____/ \____ |     "
echo "               \/             \/                          \/     "
echo "                                                                 "
echo ""
echo "    Release New Version of Cloudy and/or Update Cloudy Users"
echo ""

# Ensure we are on the master branch
currentBranch=$(git rev-parse --abbrev-ref HEAD)
if [[ $currentBranch != "master" ]]; then
    echo "Warning: You are attempting to release from a branch other than 'master'. Current branch: '$currentBranch'"
    echo "Type 'proceed' to confirm you want to proceed:"
    read -r confirm
    if [[ $confirm != "proceed" ]]; then
        echo "Incorrect response. Aborting release."
        exit 1
    fi
fi

# Determine the version to update
versionIndicator='none'
upgrade="yes"
if [[ -z $1 ]]; then
    versionIndicator='patch'
    echo "Defaulting to 'patch' version upgrade"
else
    if [[ $1 == 'update' ]]; then 
        echo "Not releasing"
    elif [[ $1 != 'minor' ]] && [[ $1 != 'major' ]] && [[ $1 != 'patch' ]]; then
        echo "Invalid version $1, must be either major, minor or patch"
        exit 1
    fi 
    versionIndicator=$1
fi

if [[ -z $2 ]]; then
    upgrade='yes'
    echo "Defaulting to update clients"
else
    if [[ $2 == 'no' ]]; then 
        echo "Not updating clients"
        upgrade="no"
    elif [[ $2 == 'yes' ]]; then 
        echo "Updating clients"
        upgrade="yes"
    else
        echo "Invalid update option, must be either 'yes' or 'no'"
        exit 1
    fi 
    versionIndicator=$1
fi

# Get the latest version from GIT
current=$(git tag | sort -r --version-sort | head -n1)
echo "Current version is $current"

if [[ $versionIndicator != 'update' ]]; then
    major=$(versions "$current" 0)
    minor=$(versions "$current" 1)
    patch=$(versions "$current" 2)

    debug "Current Segments $major $minor $patch"

    if [[ $versionIndicator == 'major' ]]; then
        major=$((major+1))
        minor=0
        patch=0
    elif [[ $versionIndicator == 'minor' ]]; then
        minor=$((minor+1))    
        patch=0
    else 
        patch=$((patch+1))
    fi
    debug "Planned Segments $major $minor $patch"

    mustBeNumber "$major"
    mustBeNumber "$minor"
    mustBeNumber "$patch"
    nextVersion="v$major.$minor.$patch"

    echo "Next version: $nextVersion"
    read -p "Do you want to tag this version? (y/n): " confirmTag
    if [[ $confirmTag != "y" ]]; then
        echo "Tagging aborted."
        exit 1
    fi

    # Tag the git repo
    git tag "$nextVersion"
    git push origin "$nextVersion"
else 
    nextVersion=$current
fi

if [[ $upgrade == 'yes' ]]; then

    updatecloudy cloudy-msgraph
    updatecloudy cloudy-azure
    updatecloudy cloudy-pg
    # updatecloudy cloudy-aws
    # updatecloudy cloudy-gcp
    updatecloudy cloudy-elastic
    # updatecloudy go-arkloud
    # updatecloudy user-api
    # updatecloudy folders-api
    # updatecloudy cac-api
    updatecloudy cloudy-keycloak
    updatecloudy arkloud-api

else 
    echo "Client update skipped"
fi

echo "Done"