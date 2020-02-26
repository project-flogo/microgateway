<!--
title: Obfuscate Activity
weight: 4618
-->

# Obfuscate JSON Payload
This activity allows you to obfuscate the required value in the playload of type JSON using defined function.

Eg . If the payload has a field contianing sensitive information; this activity would obfuscate that field.
```{
    ...
    "BookingCreditCard":"41462917261957261",
    ...
    }
    becomes 
    {
       ...
       "BookingCreditCard":"**********7261",
       ... 
    }
```
## Installation

### Flogo CLI
```bash
flogo install github.com/microgateway/activity/obfuscatejson
```

## Configuration

### Settings:
| Name          | Type   | Description
|:---           | :---   | :---     
| operation     | string | The operation to perform (Allowed values are setLastFour) - **REQUIRED**
| fields        | array  | The fields of json to obfuscate - **REQUIRED**

### Supoorted Operations
| Name          | Description
|:---           | :---     
| setLastFour   | This operation adds "*" in place of characters of the field except last four.

### Input:
| Name        | Type   | Description
|:---         | :---   | :---     
| payload     | string | The message to obfuscate