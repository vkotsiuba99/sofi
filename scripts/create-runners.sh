#!/bin/bash

# Creates a group of users for managing privileges and permissions.
groupadd runners
# Changes the group for runners to the languages directory.
chgrp -R runners /sofi/languages/

# Manages control of users running any sort of script.
# Removes read, write, and execute permission for the runners group.
chmod g-rwx /sofi/languages/
# Add execute privilege for the runners group.
chmod g+x /sofi/languages/