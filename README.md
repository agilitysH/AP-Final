# AP-Final (Mini Moodle)

## Overview
Mini Moodle is a learning service for managing courses and student progress. The server is written in Go (net/http), and data is stored in MongoDB.

## Architecture
- Go `net/http` with route patterns (e.g., `GET /courses/{id}`).
- MongoDB connection via `mongo-driver`.
- Cookie-based authentication (`session_token`) and middleware.
- Indexes are created at startup via `db.EnsureIndexes`.

## Courses Collection Schema (embedded modules/items)
```
{
  _id: ObjectId,
  title: string,
  description: string,
  category: string,
  teacherId: ObjectId,
  modules: [
    {
      _id: ObjectId,
      title: string,
      order: number,
      items: [
        {
          _id: ObjectId,
          type: string,
          title: string,
          maxScore: number,
          order: number
        }
      ]
    }
  ],
  createdAt: Date,
  updatedAt: Date
}
```

## Enrollments Collection Schema
```
{
  _id: ObjectId,
  userId: ObjectId,
  courseId: ObjectId,
  status: string,
  enrolledAt: Date,
  lastAccessAt: Date
}
```

## Progress Collection Schema
```
{
  _id: ObjectId,
  userId: ObjectId,
  courseId: ObjectId,
  itemId: ObjectId,
  status: "not_started" | "in_progress" | "done",
  score: number,
  attempts: number,
  updatedAt: Date
}
```

## /me/progress Aggregation Pipeline
Pipeline (runs on `enrollments` collection with fields `userId`, `courseId`, `status`, `enrolledAt`):
```
[
  { $match: { userId: ObjectId("<userId>") } },
  {
    $lookup: {
      from: "courses",
      localField: "courseId",
      foreignField: "_id",
      as: "course"
    }
  },
  { $unwind: "$course" },
  {
    $lookup: {
      from: "progress",
      let: { courseId: "$courseId", userId: "$userId" },
      pipeline: [
        {
          $match: {
            $expr: {
              $and: [
                { $eq: ["$courseId", "$$courseId"] },
                { $eq: ["$userId", "$$userId"] }
              ]
            }
          }
        }
      ],
      as: "progress"
    }
  },
  {
    $addFields: {
      itemsCount: {
        $sum: {
          $map: {
            input: "$course.modules",
            as: "m",
            in: { $size: { $ifNull: ["$$m.items", []] } }
          }
        }
      },
      doneCount: {
        $size: {
          $filter: {
            input: "$progress",
            as: "p",
            cond: { $eq: ["$$p.status", "done"] }
          }
        }
      },
      avgScore: { $ifNull: [{ $avg: "$progress.score" }, 0] }
    }
  },
  {
    $addFields: {
      completionRate: {
        $cond: [
          { $gt: ["$itemsCount", 0] },
          { $divide: ["$doneCount", "$itemsCount"] },
          0
        ]
      }
    }
  },
  {
    $project: {
      _id: 0,
      courseId: "$course._id",
      courseTitle: "$course.title",
      itemsCount: 1,
      doneCount: 1,
      completionRate: 1,
      avgScore: 1,
      enrollmentStatus: "$status",
      enrolledAt: "$enrolledAt"
    }
  },
  { $sort: { completionRate: -1, courseTitle: 1 } }
]
```

## API Endpoints
| Method | Path | Description | Auth |
|---|---|---|---|
| POST | `/register` | Create user account | No |
| POST | `/login` | Login and set cookie | No |
| GET | `/courses?search=&category=&teacherId=&page=&limit=&sort=` | List courses with filters, pagination, sorting | No |
| POST | `/courses` | Create course (embedded modules/items allowed) | Yes |
| GET | `/courses/{id}` | Get course by id | No |
| PATCH | `/courses/{id}` | Update course fields | Yes |
| DELETE | `/courses/{id}` | Delete course | Yes |
| POST | `/courses/{id}/modules` | Add module to course (`$push`) | Yes |
| PATCH | `/courses/{id}/modules/{moduleId}` | Update module (`arrayFilters` + `$set`) | Yes |
| DELETE | `/courses/{id}/modules/{moduleId}` | Remove module (`$pull`) | Yes |
| PUT | `/courses/{courseId}/items/{itemId}/progress` | Upsert progress (status/score/attempts) | Yes |
| GET | `/me/progress` | Aggregated progress by enrollments | Yes |
| POST | `/enrollments` | Enroll current user in a course | Yes |
| GET | `/enrollments/my` | List current user's enrollments | Yes |
| DELETE | `/enrollments/{id}` | Delete own enrollment by id | Yes |
| DELETE | `/enrollments?courseId=<id>` | Delete enrollments by course (teacher only) | Yes |

## Indexes (created at startup)
- `users`: unique index on `username`.
- `enrollments`: unique compound index on `{ userId: 1, courseId: 1 }`.
- `enrollments`: compound index on `{ courseId: 1, status: 1 }`.
- `progress`: unique compound index on `{ userId: 1, courseId: 1, itemId: 1 }`.

## UI Pages
- `/courses` � course catalog (search/filters/pagination via API)
- `/courses/{id}` � course details + modules/items list + update progress
