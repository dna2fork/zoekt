# Security Notes

- How to deal with packet dup?
  - A --[send req(encrypted er1)]--> B
  - C (get er1) --[send er1]--> B (dup action!)
  - TODO: add timestamp and more metainfo in auth token
- How to deal with large file view?
  - TODO: limit max file size for viewing
- How to deal with injection attack?
  - TODO: check user input before executing git/p4 cmd
