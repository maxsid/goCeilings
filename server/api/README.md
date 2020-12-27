API Commands Description
========================
## 1. Login requests
`POST /login` - request for getting of JWT key. 
*Request body*:
```json
{"login": "maxim", "password": "123456789"}
```
*Response* will be contained JWT token in the next view:
```json
{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJnaWZ0IjoiQ29va2llISJ9.W3AVstDky1enma2NBQ5fkryr_iWJDV-DU-OOHFl7dLM"}
```
This token has to use in the `Authorization` header value (like this: Authorization: Bearer <JWT_KEY>). 
Token is valid for 1 day. Then you have to request a new one with `POST /login` method.
#### Users
All request `users` requests require Admin role (field role=1) 

----------------------------------------------------
`GET /users?p={p}&=lim={lim}` - get users list.
This request may have the next variables: 
- `p` is a page of the list 
- `lim` is a limit of user records on one page. 
Both parameters aren't necessary and by default equal `p` = 1 and `lim` = 30. 
*Response*:
```json
{
    "users": [
        {
            "id": 1,
            "login": "admin",
            "role": 1
        },
        {
            "id": 2,
            "login": "maxim",
            "role": 2
        },
        {
            "id": 3,
            "login": "elena",
            "role": 2
        }
    ],
    "amount": 3,
    "page": 1,
    "page_limit": 30,
    "pages": 1
}
```
+ `users` is a list of users where:
    + `id` - user ID
    + `login` - login for authorization.
    + `role` - access level. At this moment `role=1` for admins, `role=2` for users.
+ `amount` - amount of the users in a database.
+ `page` - current page of the list (p).
+ `page_limit` - a number of users on one page (lim)
+ `pages` - a number of the pages of list.

-------------------------------
`POST /users` - create users.
*Request body*:
```json
{"login": "maxim", "password": "123456789", "role": 2}
```
Description of fields see at `GET /users` chapter.
Fields `login` and `password` is required. Role by default is equal 2.
*Response*: If the response has code 201, then the request has been completed successfully.

-----------------------------------------------------------------
`GET /users/{id}` - get information about one user by ID. 
*Response*:
```json
{"id": "1", "login": "maxim", "role": 2}
```
----------------------------------------------------
`PUT /users/{id}` - update information about a user by ID.
*Request*: body for the login changing of the user:
```json
{"login": "maxim", "password": "123", "role": 2}
```
Not necessary specify all fields. Only specified ones will be updated.
*Response*: If the response has code 200, then the request has been completed successfully.
-----------------------------------------
`DELETE /users/{id}` - delete a user by ID.
*Response*: If the response has code 200, then the request has been completed successfully.
-----------------------------------------
## 2. Drawing requests
`GET /drawings?p={p}&=lim={lim}` - get drawings accessible for current user. 
About parameters `p` and `lim` read `GET /users`.
*Response*:
```json
{
    "drawings": [
        {
            "id": 1,
            "name": "drawing 1"
        },
        {
            "id": 2,
            "name": "Lenin st., 25"
        },
        {
            "id": 3,
            "name": "Karl Marx st., 123"
        }
    ],
    "amount": 3,
    "page": 1,
    "page_limit": 30,
    "pages": 1
}
```
Response contains next fields:
+ `drawings` - the list of drawings capable for a current user.
    + `id` - drawing id
    + `name` - drawing name  

`amount`,`page`,`page_limit` and `pages` are default parameters for all lists. About them could read in `GET /users`.

----------------------------------------
`POST /drawings` - create a new drawing. 
*Request*:
```json
{
    "name": "drawing 1",
    "points": [
        {},
        {"x": 0, "y": 125},
        {"distance": 27, "direction": 0},
        {"distance": 46, "angle": 270}
    ],
    "measures": {
        "length": "cm",
        "area": "m2",
        "perimeter": "m",
        "angle": "deg"
    }
}
```
+ `name` - a name of the new drawing. It's a required field.
+ `points` - its points. then  Can be several variants:
    + Empty - `{}` - It's equivalent `{"x": 0, "y": 0}`
    + Coordinates - `{"x": 0, "y": 125}` - add a point by coordinates
    + Direction - `{"distance": 27, "direction": 0}` - add a point at a distance and in the direction from the previous point.
     Direction is moving counterclockwise on a full circle, where `0deg (or 360deg) direction to the right`, `90deg to the up`,
     `180deg to the left` and `270deg to the down`. Require that drawing has not less one point.
     + Angle - `{"distance": 46, "angle": 270}` - add a point which will have created angle with a previous segment.
     Angle `90deg` always will have created a right angle. Require that drawing has not less two points.
+ `mesures` - a list of measures for this drawing.
    + `lenght` - can be `cm`, `mm`, `dm`, `m`, `km`, `yd`, `in`, `mi` or `ft`. Default value is `cm`.
    + `area` - can be `m2`, `cm2`, `mm2`, `dm2`, `km2`, `yd2`, `in2`, `mi2` or `ft2`. Default value is `m2`.
    + `perimeter` - measure for displaying the perimeter of the drawing. Can be the same values as the length field, but default value is `m`.
    + `angle` - can be `deg` or `rad`. Default is `deg`. 
*Response*: If the response has code 201, then the request has been completed successfully.
------------------------------------------------------
`GET /drawings/{id}` - get info about drawing by ID.
*Response*:
```json
{
    "id": 2,
    "name": "drawing 1",
    "area": 0.11,
    "perimeter": 3.71,
    "points_count": 4,
    "width": 27,
    "height": 171,
    "points": [
        {"x": 0, "y": 0},
        {"x": 0, "y": 125},
        {"x": 27, "y": 125},
        {"x": 27, "y": 171}
    ],
    "measures": {
        "length": "cm",
        "area": "m2",
        "perimeter": "m",
        "angle": "deg"
    }
}
```
+ `area` and `perimeter` - default parameters of any geometric shape.
+ `points_count` - a number of points
+ `width` - distance between the leftest point and the rightest one.
+ `height` - distance between the lowest point and the highest one.
+ `points` - all points
+ `measures` - look at `POST /drawings`

------------------------------------------------------
`DELETE /drawings/{id}` - delete drawing by its ID.
*Response*: If the response has code 200, then the request has been completed successfully.

------------------------------------------------------
`GET /drawings/{id}/points?m=cm&p=2` - get all points of the drawing.
`m` is an unnecessary parameter of the measure of coordinates. 
Can be  **m, dm, cm, mm, km, yd, in, mi or ft** (default values is **cm**)
`p` is an unnecessary parameter of precision (number of digits after dot) of coordinates.
Default value is 2. 
*Response*:
```json
{
    "id": 2,
    "name": "drawing 1",
    "points": [
        {"x": 0, "y": 0},
        {"x": 0, "y": 125},
        {"x": 27, "y": 125},
        {"x": 27, "y": 171}
    ],
    "measure": "cm"
}
```
------------------------------------------------------
`POST /drawings/{id}/points` - add points into drawing.
*Request*:
```json
{
    "points": [
        {},
        {"x": 0, "y": 125},
        {"distance": 27, "direction": 0},
        {"distance": 46, "angle": 270}
    ],
    "measures": {
        "length": "cm",
        "area": "m2",
        "perimeter": "m",
        "angle": "deg"
    }
}
```
The same JSON as in `POST /drawings`

-----------------------------------
`GET /drawings/{id}/points/{n}?m=cm&p=2` - get point coordinates.
Parameter `m` is length measure and `p` is a number of digits after dot.

*Response* example:
```json
{
    "x": 0,
    "y": 49.21,
    "measure": "in"
}
```
-------------------
`DELETE /drawings/{id}/points/{n}` - delete point of the drawing by position.
*Response*: If the response has code 200, then the request has been completed successfully.
-------------------
`PUT /drawings/{id}/points/{n}` - update point.
Point record has the same rule as `POST /drawings` and `POST /drawings/{id}/points`, but only one point.
 ```json
 {
     "point": {"distance": 27, "direction": 0},
     "measures": {
         "length": "cm",
         "area": "m2",
         "perimeter": "m",
         "angle": "deg"
     }
 }
 ```
-------------------
`GET /drawings/{id}/image?info=true` - get png image of the drawing.
If parameter `info=true` then in the image will be included information about 
drawing, like area, perimeter, width and other. 

## 3. Drawing permissions management

All users in the database have a role of `admin` or `user`. `Admin` has all permissions for any requests, including `/users`
and even requests on behalf of any other user. `User`s have only next:
- Get user information about himself or herself.
- Create new drawings.
- Get/Change/Delete/Share their own drawings.
- Get/Change/Delete/Share drawings of other users if the user has permission on this operation, which was got from 
  a drawing creator or from the user with Share permissions.
  
Permissions `Get`, `Change` and `Delete` speak for themselves. Permission `Share` allows the user to control permissions 
for other users on the drawing. This chapter provides information about permission management requests. 
`Owner` permission means that a user is the drawing creator, thus `Owner` is equivalent to `Get`, `Change`, `Delete` and 
`Share`, but cannot be deleted without the drawing.
----
`GET /users/{id}/permissions` - get all permissions of the user by ID.

*Response example*:
```json
[
  {
    "user": {
      "id": 1,
      "login": "maxim",
      "role": 1
    },
    "drawing": {
      "id": 2,
      "name": "Drawing 2"
    },
    "owner": true
  },
  {
    "user": {
      "id": 1,
      "login": "maxim",
      "role": 1
    },
    "drawing": {
      "id": 7,
      "name": "Drawing 7"
    },
    "get": true,
    "change": true,
    "delete": true,
    "share": true
  }
]
```
Not specified fields mean false.

----
`GET /drawings/{id}/permissions` - get all permissions on this drawing by ID.

*Response* has the same format as in the previous example.

---
`POST /drawings/{id}/permissions` or `POST /users/{id}/permissions` - create a permission.
The both requests do the same thing, but the first specify the drawing, and the second specify the user, so it's 
no need writing their in the request body.

*Request example*:
```json
{
  "user_id": 2,
  "get": true
}
```
or
```json
{
  "drawing_id": 3,
  "get": true,
  "change": true
}
```
Success *response* returns 201 status code.

---
`GET /users/{id}/permissions/drawings/{id}` or `GET /drawings/{id}/permissions/users/{id}` - get a separated permission
by user_id and drawing_id.

*Response example*:
```json
{
  "user": {
    "id": 1,
    "login": "maxim",
    "role": 1
  },
  "drawing": {
    "id": 5,
    "name": "Drawing 5"
  },
  "get": true,
  "change": true
}
```
---
`DELETE /users/{id}/permissions/drawings/{id}` or `DELETE /drawings/{id}/permissions/users/{id}` - 
delete a separated permission by user_id and drawing_id.

Success *response* returns 200 status code.