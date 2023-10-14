# FileQueueHTTP

FileQueueHttp is a service designed to simplify detecting and queueing files, providing a seamless solution for detecting new files on disk, automatically saving them into a FIFO (First-In-First-Out) queue, and offering a simple web server API for fetching these files. 

Here are the key features:

### Automatic File Detection:
FileQueueHttp uses inotify, a Linux kernel feature, to continuously monitor a specified folder/folders for the creation of new files.

### FIFO Queue:
Newly detected files are enqueued into a FIFO queue, ensuring that files are processed in the order they are received.

### Web Server Interface:
FileQueueHttp provides a straightforward web server interface for easy interaction with the queued files.
Users can pull files from the queue via HTTP requests, simplifying file retrieval.

### Successful File Handling:
When a file is fetched successfully through the web API, it is automatically removed  from the disk.

### Disk Space Management:
FileQueueHttp includes disk space management features.
You can configure a maximum disk space limit to prevent storage overuse.
The system will automatically prune files (remove and delete) from the queue and disk to stay within the specified limit.

### Built-in Prometheus Metrics:
FileQueueHttp offers built-in Prometheus metrics, allowing you to monitor the application's performance and file management statistics.

Key metrics include:
* Number of files in the queue.
* Number of fetched files.
* Number of files unable to fetch (typically due to pre-deletion).
* Number of files pruned due to reaching the disk space limit.
