-- Seeds 50 problems (spread one/day over the past 50 days) with 0-3 random
-- submissions each, for exercising the History screen locally.
-- Run: mysql -h127.0.0.1 -P3306 -udailycodingcompanion -p dailycodingcompanion < seed_history_data.sql

DELIMITER $$

DROP PROCEDURE IF EXISTS seed_history_data$$

CREATE PROCEDURE seed_history_data(IN target_user_id CHAR(36))
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE j INT DEFAULT 0;
    DECLARE sub_count INT DEFAULT 0;
    DECLARE new_problem_id CHAR(36);
    DECLARE new_solution_id CHAR(36);
    DECLARE problem_created TIMESTAMP;

    WHILE i <= 50 DO
        SET new_problem_id = UUID();
        SET problem_created = NOW() - INTERVAL (50 - i) DAY;

        INSERT INTO problems (problem_id, user_id, raw_problem, title, problem_text, algorithm_tag, difficulty, ai_help, needs_review_flag, created_at, updated_at)
        VALUES (
            new_problem_id,
            target_user_id,
            CONCAT('Raw ingested email body for problem ', i),
            ELT(i,
                'Two Sum', 'Add Two Numbers', 'Longest Substring Without Repeating Characters',
                'Median of Two Sorted Arrays', 'Longest Palindromic Substring', 'Zigzag Conversion',
                'Reverse Integer', 'String to Integer (atoi)', 'Container With Most Water',
                'Integer to Roman', 'Merge Intervals', '3Sum', 'Letter Combinations of a Phone Number',
                'Remove Nth Node From End of List', 'Valid Parentheses', 'Merge Two Sorted Lists',
                'Generate Parentheses', 'Merge k Sorted Lists', 'Swap Nodes in Pairs',
                'Next Permutation', 'Longest Valid Parentheses', 'Search in Rotated Sorted Array',
                'Find First and Last Position of Element in Sorted Array', 'Combination Sum',
                'Trapping Rain Water', 'Permutations', 'Rotate Image', 'Group Anagrams',
                'Maximum Subarray', 'Jump Game', 'Merge Sorted Array', 'Unique Paths',
                'Minimum Path Sum', 'Climbing Stairs', 'Word Search', 'Largest Rectangle in Histogram',
                'Binary Tree Inorder Traversal', 'Validate Binary Search Tree',
                'Symmetric Tree', 'Binary Tree Level Order Traversal', 'Maximum Depth of Binary Tree',
                'Construct Binary Tree from Preorder and Inorder Traversal', 'Flatten Binary Tree to Linked List',
                'Best Time to Buy and Sell Stock', 'Word Ladder', 'Longest Consecutive Sequence',
                'Number of Islands', 'Course Schedule', 'Coin Change', 'Kth Largest Element in an Array',
                'LRU Cache'),
            CONCAT('Full problem statement text for problem ', i, ' goes here.'),
            ELT(FLOOR(1 + RAND() * 6), 'Array', 'Graph', 'Dynamic Programming', 'Tree', 'String', 'Greedy'),
            ELT(FLOOR(1 + RAND() * 3), 'Easy', 'Medium', 'Hard'),
            IF(RAND() < 0.5,
                JSON_OBJECT(
                    'concept', 'Placeholder concept explanation',
                    'nudge', 'Placeholder nudge',
                    'approach', 'Placeholder approach',
                    'walkthrough', 'Placeholder walkthrough'
                ),
                NULL),
            RAND() < 0.15,
            problem_created,
            problem_created
        );

        SET sub_count = FLOOR(RAND() * 4); -- 0..3 submissions
        SET j = 0;
        WHILE j < sub_count DO
            SET new_solution_id = UUID();
            INSERT INTO submitted_solutions (solution_id, problem_id, solution, status, submitted_at)
            VALUES (
                new_solution_id,
                new_problem_id,
                CONCAT('# attempt ', j + 1, '\ndef solve():\n    pass'),
                IF(RAND() < 0.4, 'Solved', 'Failed'),
                problem_created + INTERVAL j HOUR
            );
            SET j = j + 1;
        END WHILE;

        SET i = i + 1;
    END WHILE;
END$$

DELIMITER ;

CALL seed_history_data('01a01586-6593-73bf-86d8-1c138d20544c');

DROP PROCEDURE seed_history_data;
