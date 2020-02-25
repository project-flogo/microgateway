<!--
title: Obfuscate Activity
weight: 4618
-->

# Obfuscate
This activity allows you to obfuscate the required value in the JSON using defined function.

## Installation

### Flogo CLI
```bash
flogo install github.com/microgateway/activity/obfuscate
```

## Configuration

### Settings:
| Name          | Type   | Description
|:---           | :---   | :---     
| operation     | string | The operation to perfirm (Allowed values are setLastFour) - **REQUIRED**
| fields        | array  | The fields of json to obfuscate - **REQUIRED**

### Input:
| Name        | Type   | Description
|:---         | :---   | :---     
| payload     | string | The message to obfuscate