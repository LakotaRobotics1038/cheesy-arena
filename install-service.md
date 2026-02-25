# To install Cheesy Arena as a service, run the following commands

```bash
ln -s /opt/apps/cheesy-arena/cheesy-arena.service /etc/systemd/system

systemctl daemon-reload
systemctl enable cheesy-arena
systemctl start cheesy-arena
```
